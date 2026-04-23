package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/masudur-rahman/styx/dberr"
	isql "github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/postgres/lib"
	"github.com/masudur-rahman/styx/validation"
)

type Postgres struct {
	conn      *sql.DB
	tx        *sql.Tx
	statement lib.Statement
}

func NewPostgres(conn *sql.DB) Postgres {
	return Postgres{conn: conn}
}

var _ isql.Engine = Postgres{}

func (pg Postgres) BeginTx(ctx context.Context) (isql.Engine, error) {
	if pg.tx != nil {
		return nil, dberr.ErrTransactionAlreadyStarted
	}
	tx, err := pg.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	pg.tx = tx
	return pg, nil
}

func (pg Postgres) Commit() error {
	if pg.tx == nil {
		return dberr.ErrTransactionNotStarted
	}
	err := pg.tx.Commit()
	pg.tx = nil
	return err
}

func (pg Postgres) Rollback() error {
	if pg.tx == nil {
		return dberr.ErrTransactionNotStarted
	}
	err := pg.tx.Rollback()
	pg.tx = nil
	return err
}

func (pg Postgres) Table(name string) isql.Engine {
	pg.statement.Table(name)
	return pg
}

func (pg Postgres) ID(id any) isql.Engine {
	pg.statement.ID(id)
	return pg
}

func (pg Postgres) In(col string, values ...any) isql.Engine {
	pg.statement.In(col, values...)
	return pg
}

func (pg Postgres) Where(cond string, args ...any) isql.Engine {
	pg.statement.Where(cond, args...)
	return pg
}

func (pg Postgres) Columns(cols ...string) isql.Engine {
	pg.statement.Columns(cols...)
	return pg
}

func (pg Postgres) AllCols() isql.Engine {
	pg.statement.AllCols()
	return pg
}

func (pg Postgres) MustCols(cols ...string) isql.Engine {
	pg.statement.MustCols(cols...)
	return pg
}

func (pg Postgres) MustFilterCols(cols ...string) isql.Engine {
	pg.statement.MustFilterCols(cols...)
	return pg
}

func (pg Postgres) ShowSQL(showSQL bool) isql.Engine {
	pg.statement.ShowSQL(showSQL)
	return pg
}

func (pg Postgres) OrderBy(col string, direction ...string) isql.Engine {
	pg.statement.OrderBy(col, direction...)
	return pg
}

func (pg Postgres) Limit(n int64) isql.Engine {
	pg.statement.Limit(n)
	return pg
}

func (pg Postgres) Offset(n int64) isql.Engine {
	pg.statement.Offset(n)
	return pg
}

func (pg Postgres) Distinct() isql.Engine {
	pg.statement.Distinct()
	return pg
}

func (pg Postgres) GroupBy(cols ...string) isql.Engine {
	pg.statement.GroupBy(cols...)
	return pg
}

func (pg Postgres) Having(cond string, args ...any) isql.Engine {
	pg.statement.Having(cond, args...)
	return pg
}

func (pg Postgres) Or(cond string, args ...any) isql.Engine {
	pg.statement.Or(cond, args...)
	return pg
}

func (pg Postgres) Like(col string, pattern string) isql.Engine {
	pg.statement.Like(col, pattern)
	return pg
}

func (pg Postgres) NotLike(col string, pattern string) isql.Engine {
	pg.statement.NotLike(col, pattern)
	return pg
}

func (pg Postgres) Exists(subquery string, args ...any) isql.Engine {
	pg.statement.Exists(subquery, args...)
	return pg
}

func (pg Postgres) NotExists(subquery string, args ...any) isql.Engine {
	pg.statement.NotExists(subquery, args...)
	return pg
}

func (pg Postgres) Count(col string, alias ...string) isql.Engine {
	pg.statement.Count(col, alias...)
	return pg
}

func (pg Postgres) Sum(col string, alias ...string) isql.Engine {
	pg.statement.Sum(col, alias...)
	return pg
}

func (pg Postgres) Avg(col string, alias ...string) isql.Engine {
	pg.statement.Avg(col, alias...)
	return pg
}

func (pg Postgres) Min(col string, alias ...string) isql.Engine {
	pg.statement.Min(col, alias...)
	return pg
}

func (pg Postgres) Max(col string, alias ...string) isql.Engine {
	pg.statement.Max(col, alias...)
	return pg
}

func (pg Postgres) Paginate(page, perPage int64) isql.Engine {
	pg.statement.Paginate(page, perPage)
	return pg
}

func (pg Postgres) Join(table, condition string) isql.Engine {
	pg.statement.Join(table, condition)
	return pg
}

func (pg Postgres) LeftJoin(table, condition string) isql.Engine {
	pg.statement.LeftJoin(table, condition)
	return pg
}

func (pg Postgres) RightJoin(table, condition string) isql.Engine {
	pg.statement.RightJoin(table, condition)
	return pg
}

func (pg Postgres) InnerJoin(table, condition string) isql.Engine {
	pg.statement.InnerJoin(table, condition)
	return pg
}

func (pg Postgres) EnableValidation(enable bool) isql.Engine {
	pg.statement.EnableValidation(enable)
	return pg
}

func (pg Postgres) WithDeleted() isql.Engine {
	pg.statement.WithDeleted()
	return pg
}

// detectSoftDelete sets soft delete column from struct tags if present.
func (pg Postgres) detectSoftDelete(doc any) Postgres {
	if col := isql.ExtractSoftDeleteColumn(doc); col != "" {
		pg.statement.SoftDeleteCol(col)
	}
	return pg
}

func (pg Postgres) ForceDelete(ctx context.Context, filter ...any) error {
	pg.statement.SetForceDelete()
	return pg.DeleteOne(ctx, filter...)
}

func (pg Postgres) Restore(ctx context.Context, filter ...any) error {
	pg.statement.GenerateWhereClause(filter...)
	if err := pg.statement.CheckWhereClauseNotEmpty(); err != nil {
		return err
	}

	query := pg.statement.GenerateRestoreQuery()
	result, err := pg.statement.ExecuteWriteQuery(ctx, pg.conn, pg.tx, query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return dberr.ErrNotFound
	}
	return nil
}

func (pg Postgres) FindOne(ctx context.Context, document any, filter ...any) (bool, error) {
	pg = pg.detectSoftDelete(document)
	pg.statement.GenerateWhereClause(filter...)

	if err := pg.statement.CheckWhereClauseNotEmpty(); err != nil {
		return false, err
	}

	query := pg.statement.GenerateReadQuery(document)
	err := pg.statement.ExecuteReadQuery(ctx, pg.conn, pg.tx, query, document)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}

	return false, err
}

func (pg Postgres) FindMany(ctx context.Context, documents any, filter ...any) error {
	pg = pg.detectSoftDelete(documents)
	pg.statement.GenerateWhereClause(filter...)

	query := pg.statement.GenerateReadQuery(documents)
	return pg.statement.ExecuteReadQuery(ctx, pg.conn, pg.tx, query, documents)
}

func (pg Postgres) InsertOne(ctx context.Context, document any) (id any, err error) {
	if pg.statement.ShouldValidate() {
		if err := validation.Validate(document); err != nil {
			return nil, err
		}
	}
	pkCol := isql.GetPKColumn(document)
	pg.statement.PKColumn(pkCol)
	query := pg.statement.GenerateInsertQuery(document)
	id, err = pg.statement.ExecuteInsertQuery(ctx, pg.conn, pg.tx, query)
	if err != nil {
		return nil, err
	}
	return assignID(document, id)
}

func (pg Postgres) InsertMany(ctx context.Context, documents []any) ([]any, error) {
	var ids []any
	for _, doc := range documents {
		pkCol := isql.GetPKColumn(doc)
		pg.statement.PKColumn(pkCol)
		query := pg.statement.GenerateInsertQuery(doc)
		id, err := pg.statement.ExecuteInsertQuery(ctx, pg.conn, pg.tx, query)
		if err != nil {
			return nil, err
		}

		_, err = assignID(doc, id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func assignID(document any, id any) (any, error) {
	val := reflect.ValueOf(document)
	if val.Kind() != reflect.Ptr {
		return document, nil
		// first make it backward compatible
		// return id, fmt.Errorf("document must be a pointer to a struct")
	}

	valElem := val.Elem()
	if valElem.Kind() != reflect.Struct {
		return id, fmt.Errorf("document must be a pointer to a struct")
	}

	var idField = fetchIDField(valElem)
	if !idField.CanSet() {
		return id, fmt.Errorf("ID field is not settable")
	}

	idVal := reflect.ValueOf(id)
	if idField.Kind() == reflect.Ptr {
		elemType := idField.Type().Elem()
		if !idVal.Type().AssignableTo(elemType) && !idVal.Type().ConvertibleTo(elemType) {
			return id, fmt.Errorf("ID type %s cannot be assigned to pointer element type %s", idVal.Type(), elemType)
		}
		idValPtr := reflect.New(elemType)
		if idVal.Type().AssignableTo(elemType) {
			idValPtr.Elem().Set(idVal)
		} else {
			idValPtr.Elem().Set(idVal.Convert(elemType))
		}
		idField.Set(idValPtr)
	} else {
		if !idVal.Type().AssignableTo(idField.Type()) {
			if idVal.Type().ConvertibleTo(idField.Type()) {
				idVal = idVal.Convert(idField.Type())
			} else {
				return id, fmt.Errorf("ID type %s cannot be assigned or converted to field type %s", idVal.Type(), idField.Type())
			}
		}
		idField.Set(idVal)
	}

	return id, nil
}

func fetchIDField(valElem reflect.Value) (idField reflect.Value) {
	for i := 0; i < valElem.NumField(); i++ {
		field := valElem.Type().Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			dbTag = strings.Split(dbTag, ",")[0]
		}
		jsonTag := field.Tag.Get("json")
		if dbTag == "id" || jsonTag == "id" {
			idField = valElem.Field(i)
			return idField
		}
	}

	idFieldNames := []string{"ID", "Id"}
	for _, name := range idFieldNames {
		idField = valElem.FieldByName(name)
		if idField.IsValid() {
			return idField
		}
	}
	return
}

func (pg Postgres) UpdateOne(ctx context.Context, document any) error {
	if pg.statement.ShouldValidate() {
		if err := validation.Validate(document); err != nil {
			return err
		}
	}
	pg.statement.GenerateWhereClause()
	if err := pg.statement.CheckWhereClauseNotEmpty(); err != nil {
		return err
	}

	query := pg.statement.GenerateUpdateQuery(document)
	result, err := pg.statement.ExecuteWriteQuery(ctx, pg.conn, pg.tx, query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return dberr.ErrNotFound
	}
	return nil
}

func (pg Postgres) DeleteOne(ctx context.Context, filter ...any) error {
	if len(filter) > 0 {
		pg = pg.detectSoftDelete(filter[0])
	}
	pg.statement.GenerateWhereClause(filter...)
	if err := pg.statement.CheckWhereClauseNotEmpty(); err != nil {
		return err
	}

	var query string
	if pg.statement.IsSoftDelete() {
		query = pg.statement.GenerateSoftDeleteQuery()
	} else {
		query = pg.statement.GenerateDeleteQuery()
	}
	result, err := pg.statement.ExecuteWriteQuery(ctx, pg.conn, pg.tx, query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return dberr.ErrNotFound
	}
	return nil
}

func (pg Postgres) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return pg.conn.QueryContext(ctx, query, args...)
}

func (pg Postgres) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return pg.conn.ExecContext(ctx, query, args...)
}

func (pg Postgres) Sync(ctx context.Context, tables ...any) error {
	for _, table := range tables {
		if err := lib.SyncTable(ctx, pg.conn, table); err != nil {
			return err
		}
	}

	return nil
}

func (pg Postgres) DropTable(ctx context.Context, name string) error {
	return lib.DropTable(ctx, pg.conn, name)
}

func (pg Postgres) Close() error {
	return pg.conn.Close()
}

func (pg Postgres) cleanup() {
	pg.statement = lib.Statement{}
}
