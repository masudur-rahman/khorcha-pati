package event

import (
	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"

	isql "github.com/masudur-rahman/styx/sql"
)

type SQLEventRepository struct {
	db     isql.Engine
	logger logr.Logger
}

func NewSQLEventRepository(db isql.Engine, logger logr.Logger) *SQLEventRepository {
	return &SQLEventRepository{
		db:     db.Table(models.Event{}.TableName()),
		logger: logger,
	}
}

func (e *SQLEventRepository) AddEvent(event string) error {
	return nil
}

func (e *SQLEventRepository) ListEvents() ([]models.Event, error) {
	return nil, nil
}
