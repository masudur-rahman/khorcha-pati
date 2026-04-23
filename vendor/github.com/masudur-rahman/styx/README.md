# Styx

Database Engine for different SQL and NoSQL databases.

## Install

```shell
go get -u github.com/masudur-rahman/styx
```

## Supported Databases

| Database   | Package              | Status     |
|------------|----------------------|------------|
| SQLite     | `sql/sqlite`         | Stable     |
| PostgreSQL | `sql/postgres`       | Stable     |
| Supabase   | `sql/supabase`       | Partial    |
| ArangoDB   | `nosql/arango`       | Stable     |
| MongoDB    | `nosql/mongo`        | Stable     |

## Quickstart

Check out the [Quickstart Example](examples/quickstart.go) for a complete guide.

```go
package main

import (
	"context"
	"time"

	"github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"
)

type User struct {
	ID        int64     `db:"id,pk autoincr"`
	Name      string    `db:"name,uq"`
	FullName  string    `db:"full_name,uqs"`
	Email     string    `db:",uqs"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	ctx := context.Background()
	conn, _ := lib.GetSQLiteConnection("test.db")

	db := sqlite.NewSQLite(conn)

	// Migrate database
	db.Sync(ctx, User{})

	// Fluent CRUD
	db.Table("user").InsertOne(ctx, &User{Name: "masud", FullName: "Masudur Rahman", Email: "masud@example.com"})

	var user User
	db.Table("user").ID(1).FindOne(ctx, &user)
	db.Table("user").Where("email=?", "masud@example.com").FindOne(ctx, &user)

	db.Table("user").ID(user.ID).UpdateOne(ctx, User{Email: "test@example.com"})

	db.Table("user").ID(1).DeleteOne(ctx)
}
```

## Struct Tags

Styx uses the `db` struct tag to map Go struct fields to database columns and define schema constraints.

### Tag Format

```
db:"column_name,options"
```

- **column_name** (before the comma): Sets the database column name. If empty, the field name is converted to `snake_case` automatically.
- **options** (after the comma): Space-separated list of constraint/behavior flags.

### Available Options

| Tag        | Purpose                          | DDL Effect                                       | Query Effect |
|------------|----------------------------------|--------------------------------------------------|--------------|
| `pk`       | Primary key                      | Adds `PRIMARY KEY` constraint                    | -            |
| `autoincr` | Auto-increment                   | `INTEGER PRIMARY KEY AUTOINCREMENT` (SQLite) / `SERIAL`/`BIGSERIAL` (Postgres) | -            |
| `uq`       | Unique constraint (single column)| Adds `UNIQUE` constraint                         | -            |
| `uqs`      | Unique composite group           | Adds composite `UNIQUE(col1, col2, ...)` across all `uqs` fields | -            |
| `req`      | Required (never skip zero-value) | None                                             | Always includes the field in WHERE, INSERT, and UPDATE queries, even when zero-valued |

### Examples

```go
type Budget struct {
	ID         int64  `db:"id,pk autoincr"`     // primary key, auto-increment
	UserID     int64  `db:"user_id,uqs"`        // part of composite unique constraint
	CategoryID string `db:"category_id,uqs req"` // composite unique + required (never skipped)
	AlertAt    int64  `db:"alert_at,req"`        // required: always included even when 0
	Amount     int64  `db:"amount"`              // regular field, skipped when zero
	Label      string `db:"label,uq"`           // single-column unique constraint
}
```

### How Zero-Value Handling Works

By default, Styx skips zero-valued fields (`""`, `0`, `false`, `time.Time{}`) in:
- **WHERE clauses** (struct filters passed to `FindOne`, `FindMany`, `DeleteOne`)
- **INSERT** queries
- **UPDATE** queries

This is useful most of the time (you don't want `WHERE id=0 AND created_at='0001-01-01'`), but it causes bugs when a zero value is intentional (e.g., `CategoryID=""` means "overall budget").

There are three ways to override this:

#### 1. `req` tag (declarative, per-field)

Mark the field once in the struct definition. It applies to all operations automatically.

```go
type Budget struct {
	CategoryID string `db:"category_id,req"` // "" is always included
}

db.FindOne(&b, Budget{UserID: 99, CategoryID: ""})
// WHERE user_id=99 AND category_id=''

db.InsertOne(&Budget{UserID: 99, CategoryID: "", Amount: 500})
// INSERT INTO "budget" (user_id, category_id, amount) VALUES (99, '', 500)
```

#### 2. `MustFilterCols` (per-query, WHERE only)

Opt in per query for specific columns in WHERE clauses.

```go
db.MustFilterCols("category_id").FindOne(&b, Budget{UserID: 99, CategoryID: ""})
// WHERE user_id=99 AND category_id=''

db.MustFilterCols("category_id").DeleteOne(Budget{UserID: 99, CategoryID: ""})
// DELETE FROM "budget" WHERE user_id=99 AND category_id=''
```

#### 3. `MustCols` (per-query, INSERT/UPDATE only)

Opt in per query for specific columns in INSERT and UPDATE.

```go
db.MustCols("alert_at", "category_id").InsertOne(&budget)
// Includes alert_at and category_id even when zero
```

#### 4. `AllCols` (per-query, all fields)

Include every field regardless of zero value. Use with caution.

```go
db.AllCols().InsertOne(&budget)
// Includes all fields, including id=0, created_at=zero, etc.
```

## Engine API

All database engines implement the `sql.Engine` interface. Methods are chainable.

### Query Building

| Method                              | Description                                |
|-------------------------------------|--------------------------------------------|
| `Table(name string)`                | Set target table name                      |
| `ID(id any)`                        | Filter by primary key                      |
| `Where(cond string, args ...any)`   | Add raw WHERE condition with `?` placeholders |
| `In(col string, values ...any)`     | Add `col IN (...)` filter                  |
| `Columns(cols ...string)`           | Select specific columns (default: `*`)     |
| `OrderBy(col, dir)`                 | Sort results (`ASC` or `DESC`)             |
| `Paginate(page, perPage)`           | Automatic `LIMIT` and `OFFSET`             |
| `Join(table, on)`                   | Add `JOIN` (also `LeftJoin`, `InnerJoin`)  |
| `GroupBy(cols...)`                  | Add `GROUP BY` clause                      |
| `Having(cond, args...)`             | Add `HAVING` clause for groups             |
| `Distinct()`                        | Enable `SELECT DISTINCT`                   |

### Features

#### Aggregates
Perform calculations directly through the query builder:
```go
db.Table("user").Count("id", "total_users").FindMany(&results)
db.Table("user").Avg("age", "average_age").FindMany(&results)
// Supported: Count, Sum, Avg, Min, Max
```

#### Soft Delete
Declaratively enable soft deletes using struct tags:
```go
type User struct {
    ID        int64      `db:"id,pk"`
    DeletedAt *time.Time `db:"deleted_at,soft_delete"`
}

db.DeleteOne(User{ID: 1}) // Sets deleted_at = CURRENT_TIMESTAMP
db.FindMany(&users)       // Automatically filters out rows where deleted_at IS NOT NULL
db.WithDeleted().FindMany(&users) // Includes deleted rows
```

#### Struct Validation
Integrate validation rules into your models:
```go
type User struct {
    Email string `db:"email" validate:"required,email"`
}

db.EnableValidation(true).InsertOne(&user) // Returns error if validation fails
```

### Zero-Value Control

| Method                              | Description                                |
|-------------------------------------|--------------------------------------------|
| `AllCols()`                         | Include all fields (INSERT/UPDATE/WHERE)   |
| `MustCols(cols ...string)`          | Force specific columns in INSERT/UPDATE    |
| `MustFilterCols(cols ...string)`    | Force specific columns in WHERE clauses    |

### CRUD Operations

| Method                                          | Description                          |
|-------------------------------------------------|--------------------------------------|
| `FindOne(doc any, filter ...any) (bool, error)` | Find one record. Returns false if not found. |
| `FindMany(docs any, filter ...any) error`       | Find multiple records into a slice   |
| `InsertOne(doc any) (id any, err error)`        | Insert one record. Returns inserted ID. |
| `InsertMany(docs []any) ([]any, error)`         | Insert multiple records              |
| `UpdateOne(doc any) error`                      | Update one record (requires WHERE)   |
| `DeleteOne(filter ...any) error`                | Delete one record (requires WHERE)   |

### Transactions

```go
tx, err := db.BeginTx()
tx.Table("user").InsertOne(&user)
tx.Commit()   // or tx.Rollback()
```

### Schema Migration

```go
db.Sync(User{}, Budget{}, Wallet{})
```

Creates tables if they don't exist, adds missing columns to existing tables.

### Raw Queries

```go
rows, err := db.Query("SELECT * FROM user WHERE name = ?", "masud")
result, err := db.Exec("DELETE FROM user WHERE id = ?", 1)
```
## Unit of Work

Styx provides a Unit of Work pattern to coordinate transactions across multiple database engines (SQL + NoSQL). See [Unit of Work Documentation](docs/unit_of_work.md) for more details.

```go
uow := styx.NewUnitOfWork(sqlEngine, nosqlEngine)

err := uow.Execute(func(sqlTx sql.Engine, nosqlTx nosql.Engine) error {
	sqlTx.Table("user").InsertOne(&user)
	nosqlTx.Collection("logs").InsertOne(logEntry)
	return nil
})
```

## Project Structure

```
sql/            SQL Engine interface + implementations
  sqlite/       SQLite (via modernc.org/sqlite, pure Go)
  postgres/     PostgreSQL (direct + gRPC remote access)
  supabase/     Supabase REST-based
  mock/         Mock SQL engine (GoMock)
nosql/          NoSQL Engine interface + implementations
  arango/       ArangoDB
  mongo/        MongoDB
  mock/         Mock NoSQL engine
dberr/          Shared error types (DataNotFound, RequirementMissing)
uow.go          Unit of Work coordinator
```

---

### Why the name Styx?

In Greek mythology, the River Styx separates the world of the living from the world of the dead.
Similarly, this ORM acts as a bridge between your application code and the database,
facilitating the flow of data between the two realms while ensuring data integrity and controlled access.
