package lib

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	isql "github.com/masudur-rahman/styx/sql"

	"github.com/iancoleman/strcase"
)

type fieldInfo struct {
	Name        string
	Type        string
	IsComposite bool
}

func GenerateTableName(table interface{}) string {
	return isql.GetTableName(table)
}

func getTableInfo(table interface{}) ([]fieldInfo, error) {
	tableType := reflect.TypeOf(table)
	tableValue := reflect.ValueOf(table)

	if tableType.Kind() == reflect.Ptr {
		tableType = tableType.Elem()
		tableValue = tableValue.Elem()
	}

	if tableType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("getTableInfo: table is expected to be struct, got %v", tableType.Kind())
	}

	var fields []fieldInfo
	for i := 0; i < tableType.NumField(); i++ {
		fieldType := tableType.Field(i)
		fieldValue := tableValue.Field(i)
		if !fieldType.IsExported() {
			fmt.Println("non-exported fields: ", fieldType.Name)
			continue
		}

		field := getFieldInfo(fieldType, fieldValue)

		fields = append(fields, field)
	}

	return fields, nil
}

func createTable(ctx context.Context, conn *sql.DB, tableName string, fields []fieldInfo) error {
	query := createTableQuery(tableName, fields)
	_, err := ExecuteWriteQuery(ctx, query, conn)
	return err
}

func addMissingColumns(ctx context.Context, conn *sql.DB, tableName string, fields []fieldInfo) error {
	columns, err := getExistingColumns(ctx, conn, tableName)
	if err != nil {
		return err
	}

	missingColumns := getMissingColumns(fields, columns)
	if len(missingColumns) > 0 {
		alterQuery := generateAddColumnQuery(tableName, missingColumns)
		_, err = ExecuteWriteQuery(ctx, alterQuery, conn)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFieldInfo(fieldType reflect.StructField, fieldValue reflect.Value) fieldInfo {
	fieldName := getFieldName(fieldType)
	columnConstraint, autoincr, isComposite := getFieldConstraint(fieldType)
	if columnConstraint != "" {
		columnConstraint = " " + columnConstraint
	}
	sqlType := getSQLType(fieldValue.Type(), autoincr)
	return fieldInfo{
		Name:        fieldName,
		Type:        removeDuplicateKeyword(sqlType + columnConstraint),
		IsComposite: isComposite,
	}
}

func removeDuplicateKeyword(keyword string) string {
	pk := "PRIMARY KEY"
	count := strings.Count(keyword, pk)
	if count > 1 {
		idx := strings.Index(keyword, pk)
		keyword = keyword[:idx+1] + strings.ReplaceAll(keyword[idx+1:], pk, "")
	}
	return keyword
}

func getFieldName(fieldType reflect.StructField) string {
	return isql.GetFieldName(fieldType)
}

func getFieldConstraint(fieldType reflect.StructField) (fc string, autoincr bool, isComposite bool) {
	constraints := []string{}
	if dbTag := fieldType.Tag.Get("db"); dbTag != "" {
		tagParts := strings.Split(dbTag, ",")
		if len(tagParts) > 1 {
			for _, part := range strings.Fields(tagParts[1]) {
				switch strings.ToUpper(part) {
				case "PK":
					constraints = append(constraints, "PRIMARY KEY")
				case "UQ":
					constraints = append(constraints, "UNIQUE")
				case "UQS":
					isComposite = true
				case "AUTOINCR":
					autoincr = true
				case "REQ":
					// handled at query generation time, no DDL effect
				}
			}
		}
	}

	return strings.Join(constraints, " "), autoincr, isComposite
}

// hasReqTag checks if a struct field has the "req" option in its db tag.
func hasReqTag(field reflect.StructField) bool {
	return isql.HasReqTag(field)
}

// ExtractPKColumn returns the primary key column name from a struct's pk tag.
// Returns "id" as default if no pk tag is found.
func ExtractPKColumn(table any) string {
	tableType := reflect.TypeOf(table)
	if tableType.Kind() == reflect.Ptr {
		tableType = tableType.Elem()
	}
	if tableType.Kind() == reflect.Slice {
		tableType = tableType.Elem()
	}
	if tableType.Kind() != reflect.Struct {
		return "id"
	}

	for i := 0; i < tableType.NumField(); i++ {
		field := tableType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		parts := strings.SplitN(dbTag, ",", 2)
		if len(parts) >= 2 {
			for _, part := range strings.Fields(parts[1]) {
				if strings.ToUpper(part) == "PK" {
					colName := parts[0]
					if colName == "" {
						colName = strcase.ToSnake(field.Name)
					}
					return colName
				}
			}
		}
	}

	return "id"
}

// ExtractSoftDeleteColumn returns the column name tagged with soft_delete.
// Returns empty string if no soft delete tag is found.
func ExtractSoftDeleteColumn(table any) string {
	tableType := reflect.TypeOf(table)
	if tableType.Kind() == reflect.Ptr {
		tableType = tableType.Elem()
	}
	if tableType.Kind() == reflect.Slice {
		tableType = tableType.Elem()
	}
	if tableType.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < tableType.NumField(); i++ {
		field := tableType.Field(i)
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
				colName := parts[0]
				if colName == "" {
					colName = strcase.ToSnake(field.Name)
				}
				return colName
			}
		}
	}
	return ""
}

func getUniqueColumnGroups(t reflect.Type) [][]string {
	groups := map[int][]string{}
	groupIndex := 0
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if dbTag := field.Tag.Get("db"); dbTag != "" {
			tagParts := strings.Split(dbTag, ",")
			for _, part := range tagParts[1:] {
				if strings.ToUpper(part) == "UQS" {
					groups[groupIndex] = append(groups[groupIndex], getFieldName(field))
					groupIndex++
				}
			}
		}
	}

	result := [][]string{}
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

func getExistingColumns(ctx context.Context, conn *sql.DB, tableName string) ([]string, error) {
	var columns []string

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("pragma table_info('%v')", tableName))
	if err != nil {
		return nil, fmt.Errorf("error getting columns for table %s: %v", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var x any
		var column string
		err = rows.Scan(&x, &column, &x, &x, &x, &x)
		if err != nil {
			return nil, fmt.Errorf("error scanning column for table %s: %v", tableName, err)
		}
		columns = append(columns, column)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting columns for table %s: %v", tableName, err)
	}

	return columns, nil
}

func getMissingColumns(fields []fieldInfo, columns []string) []string {
	var missingColumns []string

	for _, f := range fields {
		if !contains(columns, f.Name) {
			missingColumns = append(missingColumns, fmt.Sprintf("%s %s", f.Name, f.Type))
		}
	}

	return missingColumns
}

func getUniqueConstraints(ctx context.Context, conn *sql.DB, tableName string) ([][]string, error) {
	query := `
	SELECT kcu.column_name
	FROM information_schema.table_constraints tc
	JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
	WHERE tc.table_name = $1 AND tc.constraint_type = 'UNIQUE'
	ORDER BY kcu.ordinal_position;
	`

	rows, err := conn.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("error getting unique constraints for table %s: %v", tableName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		err = rows.Scan(&column)
		if err != nil {
			return nil, fmt.Errorf("error scanning unique constraint for table %s: %v", tableName, err)
		}
		columns = append(columns, column)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error getting unique constraints for table %s: %v", tableName, err)
	}

	var result [][]string
	if len(columns) > 0 {
		result = append(result, columns)
	}

	return result, nil
}

func generateDropConstraintStatement(tableName string, uqConstraints [][]string) string {
	sql := fmt.Sprintf("ALTER TABLE %s ", tableName)

	var dropConstraints []string
	for i := range uqConstraints {
		dropConstraints = append(dropConstraints,
			fmt.Sprintf("DROP CONSTRAINT IF EXISTS %s_uq_%d", tableName,
				i))
	}

	sql += strings.Join(dropConstraints, ", ")

	return sql
}

func generateAddConstraintStatement(tableName string,
	uqGroups [][]string) string {

	sql := fmt.Sprintf("ALTER TABLE %s ", tableName)

	var addConstraints []string
	for i, group := range uqGroups {
		addConstraints = append(addConstraints,
			fmt.Sprintf("ADD CONSTRAINT %s_uq_%d UNIQUE(%s)",
				tableName,
				i,
				strings.Join(group,
					", ")))
	}

	sql += strings.Join(addConstraints,
		", ")

	return sql
}

func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func createTableQuery(tableName string, fields []fieldInfo) string {
	var columnDefs []string
	var compositeKeyGroup []string
	for _, field := range fields {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", field.Name, field.Type))
		if field.IsComposite {
			compositeKeyGroup = append(compositeKeyGroup, field.Name)
		}
	}

	columnSQL := strings.Join(columnDefs, ", ")
	if len(compositeKeyGroup) > 0 {
		compositeKeySQL := fmt.Sprintf("UNIQUE(%s)", strings.Join(compositeKeyGroup, ", "))
		columnSQL += ", " + compositeKeySQL
	}

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s);", tableName, columnSQL)
}

func getSQLType(fieldType reflect.Type, autoincr bool) string {
	for fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	if autoincr {
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint64:
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		}
	}

	switch fieldType.Kind() {
	case reflect.Int, reflect.Int32:
		return "INTEGER"
	case reflect.Int64, reflect.Uint64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.String:
		return "TEXT"
	case reflect.Struct:
		if fieldType == reflect.TypeOf(time.Time{}) {
			return "DATETIME"
		}
	}

	return ""
}

func tableExists(ctx context.Context, conn *sql.DB, tableName string) (bool, error) {
	tableQuery := "SELECT name FROM sqlite_master WHERE type='table' AND name=?;"
	var name string
	err := conn.QueryRowContext(ctx, tableQuery, tableName).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("error checking if table exists: %v", err)
	}

	return true, nil
}

func generateAddColumnQuery(tableName string, missingColumns []string) string {
	alterQuery := fmt.Sprintf("ALTER TABLE \"%s\" ", tableName)
	var addColumns []string
	for _, col := range missingColumns {
		addColumns = append(addColumns, fmt.Sprintf("ADD COLUMN %s", col))
	}

	alterQuery += strings.Join(addColumns, ", ")
	return alterQuery
}

// indexInfo holds parsed index metadata from struct tags.
type indexInfo struct {
	Name   string
	Cols   []string
	Unique bool
}

// extractIndexes parses idx and unique_idx tags from a struct type.
// Supports: db:"col,idx" (auto-named), db:"col,idx:my_index" (named/composite).
func extractIndexes(table any) []indexInfo {
	tableType := reflect.TypeOf(table)
	if tableType.Kind() == reflect.Ptr {
		tableType = tableType.Elem()
	}
	if tableType.Kind() != reflect.Struct {
		return nil
	}

	named := map[string]*indexInfo{}
	var unnamed []indexInfo

	for i := 0; i < tableType.NumField(); i++ {
		field := tableType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			continue
		}
		parts := strings.SplitN(dbTag, ",", 2)
		if len(parts) < 2 {
			continue
		}
		colName := parts[0]
		if colName == "" {
			colName = strcase.ToSnake(field.Name)
		}

		for _, part := range strings.Fields(parts[1]) {
			lp := strings.ToLower(part)
			if lp == "idx" {
				unnamed = append(unnamed, indexInfo{Cols: []string{colName}})
			} else if lp == "unique_idx" {
				unnamed = append(unnamed, indexInfo{Cols: []string{colName}, Unique: true})
			} else if strings.HasPrefix(lp, "idx:") {
				idxName := strings.TrimPrefix(lp, "idx:")
				if existing, ok := named[idxName]; ok {
					existing.Cols = append(existing.Cols, colName)
				} else {
					named[idxName] = &indexInfo{Name: idxName, Cols: []string{colName}}
				}
			} else if strings.HasPrefix(lp, "unique_idx:") {
				idxName := strings.TrimPrefix(lp, "unique_idx:")
				if existing, ok := named[idxName]; ok {
					existing.Cols = append(existing.Cols, colName)
				} else {
					named[idxName] = &indexInfo{Name: idxName, Cols: []string{colName}, Unique: true}
				}
			}
		}
	}

	var result []indexInfo
	for _, idx := range named {
		result = append(result, *idx)
	}
	return append(result, unnamed...)
}

func createIndexes(ctx context.Context, conn *sql.DB, tableName string, indexes []indexInfo) error {
	for i, idx := range indexes {
		unique := ""
		if idx.Unique {
			unique = "UNIQUE "
		}
		name := idx.Name
		if name == "" {
			name = fmt.Sprintf("idx_%s_%d", tableName, i)
		}
		query := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS \"%s\" ON \"%s\" (%s)",
			unique, name, tableName, strings.Join(idx.Cols, ", "))
		if _, err := ExecuteWriteQuery(ctx, query, conn); err != nil {
			return fmt.Errorf("error creating index %s: %w", name, err)
		}
	}
	return nil
}

// DropTable drops a table by name.
func DropTable(ctx context.Context, conn *sql.DB, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err := ExecuteWriteQuery(ctx, query, conn)
	return err
}

func SyncTable(ctx context.Context, conn *sql.DB, table interface{}) error {
	tableName := GenerateTableName(table)
	fields, err := getTableInfo(table)
	if err != nil {
		return err
	}

	if exist, err := tableExists(ctx, conn, tableName); err != nil {
		return err
	} else if !exist {
		if err = createTable(ctx, conn, tableName, fields); err != nil {
			return err
		}
	} else {
		if err = addMissingColumns(ctx, conn, tableName, fields); err != nil {
			return err
		}
	}

	indexes := extractIndexes(table)
	if len(indexes) > 0 {
		if err = createIndexes(ctx, conn, tableName, indexes); err != nil {
			return err
		}
	}

	return nil
}
