package styx

import (
	"context"

	"github.com/masudur-rahman/styx/nosql"
	"github.com/masudur-rahman/styx/sql"
)

// UnitOfWork represents the unit of work for coordinating transactions
type UnitOfWork struct {
	SQL   sql.Engine
	NoSQL nosql.Engine
}

// Begin starts a new transaction
func (uow UnitOfWork) Begin(ctx context.Context) (UnitOfWork, error) {
	cp := UnitOfWork{
		SQL:   uow.SQL,
		NoSQL: uow.NoSQL,
	}
	if uow.SQL != nil {
		sqlTx, err := uow.SQL.BeginTx(ctx)
		if err != nil {
			return UnitOfWork{}, err
		}
		cp.SQL = sqlTx
	}
	// For NoSQL databases, no action needed for beginning a transaction
	return cp, nil
}

// Commit commits the transaction
func (uow UnitOfWork) Commit() error {
	if uow.SQL != nil {
		if err := uow.SQL.Commit(); err != nil {
			return err
		}
	}
	// For NoSQL databases, no action needed for committing a transaction
	return nil
}

// Rollback rolls back the transaction
func (uow UnitOfWork) Rollback() error {
	if uow.SQL != nil {
		if err := uow.SQL.Rollback(); err != nil {
			return err
		}
	}
	// For NoSQL databases, no action needed for rolling back a transaction
	return nil
}
