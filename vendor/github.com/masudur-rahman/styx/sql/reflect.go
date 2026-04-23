package sql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/strcase"
)

var (
	fieldMapCache  sync.Map // map[reflect.Type]map[string]int (index of field)
	tableNameCache sync.Map // map[reflect.Type]string
	pkColumnCache  sync.Map // map[reflect.Type]string
)

// GetTableName returns the database table name for a given struct, with caching.
func GetTableName(table interface{}) string {
	t := reflect.TypeOf(table)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if name, ok := tableNameCache.Load(t); ok {
		return name.(string)
	}

	// Default name from struct
	tableName := strcase.ToSnake(t.Name())
	if tableName == "" {
		// Try to get name from Type if t.Name() is empty (can happen with some reflect types)
		tableName = strcase.ToSnake(t.String())
		// t.String() might be "sqlite_test.User", we want "User"
		parts := strings.Split(tableName, ".")
		tableName = parts[len(parts)-1]
	}

	// Check for TableName() method
	// We need a value to call the method
	val := reflect.New(t)
	if method := val.MethodByName("TableName"); method.IsValid() {
		rs := method.Call([]reflect.Value{})
		tableName = rs[0].String()
	}

	tableNameCache.Store(t, tableName)
	return tableName
}

// GetDBFieldMap returns a map of database column names to field indices for a struct, with caching.
func GetDBFieldMap(doc any) map[string]int {
	t := reflect.TypeOf(doc)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if cache, ok := fieldMapCache.Load(t); ok {
		return cache.(map[string]int)
	}

	fieldMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		colName := field.Name
		if dbTag := field.Tag.Get("db"); dbTag != "" {
			tagParts := strings.Split(dbTag, ",")
			if tagParts[0] != "" {
				colName = tagParts[0]
			}
		}
		fieldMap[strcase.ToSnake(colName)] = i
	}

	fieldMapCache.Store(t, fieldMap)
	return fieldMap
}

// GetPKColumn returns the primary key column name for a struct, with caching.
func GetPKColumn(table any) string {
	t := reflect.TypeOf(table)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if col, ok := pkColumnCache.Load(t); ok {
		return col.(string)
	}

	pkCol := "id" // default
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		parts := strings.SplitN(dbTag, ",", 2)
		if len(parts) >= 2 {
			for _, part := range strings.Fields(parts[1]) {
				if strings.ToUpper(part) == "PK" {
					if parts[0] != "" {
						pkCol = parts[0]
					} else {
						pkCol = strcase.ToSnake(field.Name)
					}
					goto found
				}
			}
		}
	}

found:
	pkColumnCache.Store(t, pkCol)
	return pkCol
}

var softDeleteCache sync.Map

// ExtractSoftDeleteColumn returns the column name tagged with soft_delete, with caching.
func ExtractSoftDeleteColumn(table any) string {
	t := reflect.TypeOf(table)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if col, ok := softDeleteCache.Load(t); ok {
		return col.(string)
	}

	softDeleteCol := ""
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		parts := strings.SplitN(dbTag, ",", 2)
		if len(parts) < 2 {
			continue
		}
		for _, part := range strings.Fields(parts[1]) {
			if strings.ToLower(part) == "soft_delete" {
				softDeleteCol = parts[0]
				if softDeleteCol == "" {
					softDeleteCol = strcase.ToSnake(field.Name)
				}
				goto found
			}
		}
	}

found:
	softDeleteCache.Store(t, softDeleteCol)
	return softDeleteCol
}

// GetFieldName returns the database column name for a struct field.
func GetFieldName(field reflect.StructField) string {
	fieldName := field.Name
	if dbTag := field.Tag.Get("db"); dbTag != "" {
		colName := strings.Split(dbTag, ",")[0]
		if colName != "" {
			fieldName = colName
		}
	}
	return strcase.ToSnake(fieldName)
}

// HasReqTag checks if a struct field has the "req" option in its db tag.
func HasReqTag(field reflect.StructField) bool {
	dbTag := field.Tag.Get("db")
	if dbTag == "" {
		return false
	}
	parts := strings.SplitN(dbTag, ",", 2)
	if len(parts) < 2 {
		return false
	}
	for _, part := range strings.Fields(parts[1]) {
		if strings.ToUpper(part) == "REQ" {
			return true
		}
	}
	return false
}

// IsZeroValue checks if a value is its type's zero value.
func IsZeroValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.IsZero()
}

// ScanRow scans a database row into a struct using cached field mapping.
func ScanRow(rows *sql.Rows, doc any) error {
	fields, err := rows.Columns()
	if err != nil {
		return err
	}

	fieldMap := GetDBFieldMap(doc)
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	scans := make([]any, len(fields))
	for i := range scans {
		scans[i] = &scans[i]
	}
	if err := rows.Scan(scans...); err != nil {
		return err
	}

	for idx, col := range fields {
		rawVal := scans[idx]
		if rawVal == nil {
			continue
		}

		fieldIdx, ok := fieldMap[col]
		if !ok {
			continue
		}

		field := val.Field(fieldIdx)
		if !field.CanSet() {
			continue
		}

		// Handle type conversion if necessary
		v := reflect.ValueOf(rawVal)
		if v.Type().AssignableTo(field.Type()) {
			field.Set(v)
		} else if field.Kind() == reflect.Ptr {
			// Handle pointer assignment
			elemType := field.Type().Elem()
			if v.Type().AssignableTo(elemType) {
				newVal := reflect.New(elemType)
				newVal.Elem().Set(v)
				field.Set(newVal)
			} else if v.Type().ConvertibleTo(elemType) {
				newVal := reflect.New(elemType)
				newVal.Elem().Set(v.Convert(elemType))
				field.Set(newVal)
			} else if elemType.String() == "time.Time" {
				// Special handling for time.Time from string
				if s, ok := rawVal.(string); ok {
					if t, err := parseTime(s); err == nil {
						newVal := reflect.New(elemType)
						newVal.Elem().Set(reflect.ValueOf(t))
						field.Set(newVal)
					}
				}
			}
		} else if field.Type().String() == "time.Time" {
			if s, ok := rawVal.(string); ok {
				if t, err := parseTime(s); err == nil {
					field.Set(reflect.ValueOf(t))
				}
			}
		} else if v.Type().ConvertibleTo(field.Type()) {
			field.Set(v.Convert(field.Type()))
		}
	}
	return nil
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time: %s", s)
}
