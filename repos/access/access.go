package access

import (
	"context"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"

	isql "github.com/masudur-rahman/styx/sql"
)

type SQLAccessRepository struct {
	settings isql.Engine
	allowed  isql.Engine
	logger   logr.Logger
}

func NewSQLAccessRepository(db isql.Engine, logger logr.Logger) *SQLAccessRepository {
	return &SQLAccessRepository{
		settings: db.Table(models.Setting{}.TableName()),
		allowed:  db.Table(models.AllowedUser{}.TableName()),
		logger:   logger,
	}
}

func (r *SQLAccessRepository) GetSetting(key string) (string, bool, error) {
	ctx := context.Background()
	var s models.Setting
	found, err := r.settings.FindOne(ctx, &s, models.Setting{Key: key})
	if err != nil || !found {
		return "", false, err
	}
	return s.Value, true, nil
}

func (r *SQLAccessRepository) SetSetting(key, value string) error {
	ctx := context.Background()
	var existing models.Setting
	found, err := r.settings.FindOne(ctx, &existing, models.Setting{Key: key})
	if err != nil {
		return err
	}
	if found {
		existing.Value = value
		return r.settings.ID(existing.ID).MustCols("value").UpdateOne(ctx, &existing)
	}
	_, err = r.settings.InsertOne(ctx, &models.Setting{Key: key, Value: value})
	return err
}

func (r *SQLAccessRepository) ListAllowedUsers() ([]models.AllowedUser, error) {
	ctx := context.Background()
	entries := make([]models.AllowedUser, 0)
	err := r.allowed.FindMany(ctx, &entries)
	return entries, err
}

func (r *SQLAccessRepository) AddAllowedUser(entry *models.AllowedUser) error {
	ctx := context.Background()
	if entry.CreatedAt == 0 {
		entry.CreatedAt = time.Now().Unix()
	}
	_, err := r.allowed.InsertOne(ctx, entry)
	return err
}

func (r *SQLAccessRepository) UpdateAllowedUser(entry *models.AllowedUser) error {
	ctx := context.Background()
	return r.allowed.ID(entry.ID).UpdateOne(ctx, entry)
}

func (r *SQLAccessRepository) RemoveAllowedUser(id int64) error {
	ctx := context.Background()
	return r.allowed.DeleteOne(ctx, models.AllowedUser{ID: id})
}
