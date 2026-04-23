package sql

import (
	"context"
	"database/sql"
)

// Engine is the unified SQL database interface. All methods return Engine to
// enable method chaining. Implementations include Postgres, SQLite, and Supabase.
type Engine interface {
	// BeginTx starts a new transaction and returns a transaction-scoped Engine.
	BeginTx(ctx context.Context) (Engine, error)
	// Commit commits the current transaction.
	Commit() error
	// Rollback aborts the current transaction.
	Rollback() error

	// Table sets the target table name for the next operation.
	Table(name string) Engine

	// ID filters by primary key value.
	ID(id any) Engine
	// In adds an IN clause for the given column.
	In(col string, values ...any) Engine
	// Where adds a parameterised WHERE condition.
	Where(cond string, args ...any) Engine
	// Columns restricts SELECT to the named columns.
	Columns(cols ...string) Engine
	// AllCols forces all columns to be included in UPDATE.
	AllCols() Engine
	// MustCols always includes the named columns even when they have zero values.
	MustCols(cols ...string) Engine
	// MustFilterCols always includes the named columns in WHERE clauses.
	MustFilterCols(cols ...string) Engine
	// ShowSQL logs the generated SQL when enabled.
	ShowSQL(showSQL bool) Engine

	// OrderBy adds an ORDER BY clause. direction defaults to ASC.
	OrderBy(col string, direction ...string) Engine
	// Limit sets the maximum number of rows returned.
	Limit(n int64) Engine
	// Offset skips the first n rows.
	Offset(n int64) Engine
	// Distinct adds a DISTINCT modifier to the SELECT.
	Distinct() Engine
	// GroupBy groups results by the named columns.
	GroupBy(cols ...string) Engine
	// Having adds a HAVING condition (used with GroupBy).
	Having(cond string, args ...any) Engine
	// Or adds an OR condition to the WHERE clause.
	Or(cond string, args ...any) Engine
	// Like adds a LIKE pattern condition.
	Like(col string, pattern string) Engine
	// NotLike adds a NOT LIKE pattern condition.
	NotLike(col string, pattern string) Engine
	// Exists adds an EXISTS subquery condition.
	Exists(subquery string, args ...any) Engine
	// NotExists adds a NOT EXISTS subquery condition.
	NotExists(subquery string, args ...any) Engine
	// Count adds a COUNT aggregate expression.
	Count(col string, alias ...string) Engine
	// Sum adds a SUM aggregate expression.
	Sum(col string, alias ...string) Engine
	// Avg adds an AVG aggregate expression.
	Avg(col string, alias ...string) Engine
	// Min adds a MIN aggregate expression.
	Min(col string, alias ...string) Engine
	// Max adds a MAX aggregate expression.
	Max(col string, alias ...string) Engine
	// Paginate sets LIMIT and OFFSET based on 1-indexed page and per-page count.
	Paginate(page, perPage int64) Engine

	// Join adds a JOIN clause.
	Join(table, condition string) Engine
	// LeftJoin adds a LEFT JOIN clause.
	LeftJoin(table, condition string) Engine
	// RightJoin adds a RIGHT JOIN clause.
	RightJoin(table, condition string) Engine
	// InnerJoin adds an INNER JOIN clause.
	InnerJoin(table, condition string) Engine

	// WithDeleted includes soft-deleted rows in query results.
	WithDeleted() Engine
	// ForceDelete permanently deletes matching rows, bypassing soft delete.
	ForceDelete(ctx context.Context, filter ...any) error
	// Restore clears the soft-delete marker on matching rows.
	Restore(ctx context.Context, filter ...any) error
	// EnableValidation turns struct tag validation on or off for write operations.
	EnableValidation(enable bool) Engine

	// FindOne retrieves a single row into document. Returns (false, nil) when not found.
	FindOne(ctx context.Context, document any, filter ...any) (bool, error)
	// FindMany retrieves all matching rows into documents (must be a pointer to a slice).
	FindMany(ctx context.Context, documents any, filter ...any) error

	// InsertOne inserts document and returns the generated primary key.
	InsertOne(ctx context.Context, document any) (id any, err error)
	// InsertMany inserts multiple documents and returns their generated primary keys.
	InsertMany(ctx context.Context, documents []any) ([]any, error)

	// UpdateOne updates the row identified by ID() with non-zero fields from document.
	UpdateOne(ctx context.Context, document any) error

	// DeleteOne deletes a single matching row (soft or hard delete depending on schema).
	DeleteOne(ctx context.Context, filter ...any) error

	// Query executes a raw SQL query and returns the result rows.
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	// Exec executes a raw SQL statement (INSERT/UPDATE/DELETE) and returns the result.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// Sync creates or alters tables to match the provided struct schemas.
	Sync(ctx context.Context, tables ...any) error
	// DropTable drops the named table from the database.
	DropTable(ctx context.Context, name string) error

	// Close releases the underlying database connection.
	Close() error
}
