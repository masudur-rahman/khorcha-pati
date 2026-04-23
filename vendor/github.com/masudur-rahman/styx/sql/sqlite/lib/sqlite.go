package lib

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/masudur-rahman/styx/dberr"
	isql "github.com/masudur-rahman/styx/sql"

	"github.com/iancoleman/strcase"

	_ "modernc.org/sqlite"
)

// GetSQLiteConnection opens a SQLite database and returns a *sql.DB connection pool.
func GetSQLiteConnection(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}

// IsZeroValue checks if a value is its type's zero value.
// Deprecated: Use dberr.IsZeroValue instead.
func IsZeroValue(value any) bool {
	return dberr.IsZeroValue(value)
}

func toDBFieldName(fieldName string) string {
	return strcase.ToSnake(fieldName)
}

func fromDBFieldName(fieldName string) string {
	return strcase.ToLowerCamel(fieldName)
}

func GenerateReadQuery(tableName string, record map[string]any) string {
	var conditions []string
	for key, val := range record {
		if isql.IsZeroValue(val) {
			continue
		}

		col, value := toColumnValue(key, val)

		operator := " = "
		if value[0] == '(' {
			operator = " IN "
		}
		condition := strings.Join([]string{col, value}, operator)
		conditions = append(conditions, condition)
	}

	var conditionString string
	query := fmt.Sprintf("SELECT * FROM \"%s\"", tableName)

	if len(conditions) > 0 {
		conditionString = " WHERE "
		conditionString += strings.Join(conditions, " AND ")
	}

	query += conditionString

	return query
}

func toColumnValue(key string, val any) (string, string) {
	key = strcase.ToSnake(key)
	value := formatValues(val)
	return key, value
}

func formatValues(val any) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
	case []string:
		return fmt.Sprintf("('%s')", strings.Join(v, "', '"))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ExecuteWriteQuery(ctx context.Context, query string, conn *sql.DB) (sql.Result, error) {
	return conn.ExecContext(ctx, query)
}
