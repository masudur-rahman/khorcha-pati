package configs

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/infra/logr"
	"github.com/masudur-rahman/expense-tracker-bot/models"
	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"
	"github.com/masudur-rahman/expense-tracker-bot/modules/google"
	"github.com/masudur-rahman/expense-tracker-bot/services/all"

	"github.com/masudur-rahman/styx"
	isql "github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/postgres"
	sqlib "github.com/masudur-rahman/styx/sql/postgres/lib"
	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"
)

// sqlDB holds a reference to the database engine for utility functions.
var sqlDB isql.Engine

func InitiateCache() {
	cache.Init(TrackerConfig.Cache)
}

func InitiateDatabaseConnection(ctx context.Context) error {
	cfg := TrackerConfig.Database
	switch cfg.Type {
	case DatabasePostgres:
		db, err := getPostgresDatabase(ctx)
		if err != nil {
			return err
		}
		return initializeSQLServices(styx.UnitOfWork{SQL: db})
	case DatabaseSQLite, "":
		if cfg.SQLite.SyncToDrive {
			if !cfg.SQLite.DisableSyncFromDrive {
				if err := google.SyncDatabaseFromDrive(); err != nil {
					return err
				}
				logr.DefaultLogger.Infof("SQLite database synced from google drive")
			}
			go google.SyncDatabaseToDrivePeriodically(TrackerConfig.Database.SQLite.SyncInterval)
		}

		db, err := getSQLiteDatabase(ctx)
		if err != nil {
			return err
		}
		return initializeSQLServices(styx.UnitOfWork{SQL: db})
	default:
		return fmt.Errorf("unknown database type")
	}
}

func getSQLiteDatabase(ctx context.Context) (isql.Engine, error) {
	conn, err := lib.GetSQLiteConnection(google.DatabasePath())
	if err != nil {
		return nil, err
	}

	return sqlite.NewSQLite(ctx, conn), nil
}

func initializeSQLServices(uow styx.UnitOfWork) error {
	sqlDB = uow.SQL
	if err := syncTables(uow.SQL); err != nil {
		return err
	}
	if err := fixNullZeroValues(uow.SQL); err != nil {
		return err
	}
	all.InitiateSQLServices(uow, logr.DefaultLogger)

	return all.GetServices().Txn.UpdateTxnCategories()
}

// fixNullZeroValues patches existing rows where styx v1.2.3 inserted NULL for zero-value fields.
func fixNullZeroValues(db isql.Engine) error {
	stmts := []string{
		`UPDATE "transaction" SET deleted_at = 0 WHERE deleted_at IS NULL`,
		`UPDATE "wallet" SET version = 0 WHERE version IS NULL`,
		`UPDATE "contacts" SET net_balance = 0 WHERE net_balance IS NULL`,
		`UPDATE "contacts" SET last_txn_timestamp = 0 WHERE last_txn_timestamp IS NULL`,
		`UPDATE "budget" SET alert_at = 80 WHERE alert_at IS NULL`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("fix null values: %w", err)
		}
	}
	return nil
}

//func getServicesForSupabase(ctx context.Context) *all.Services {
//	supClient := supabase.InitializeSupabase(ctx)
//
//	var db isql.Engine
//	db = supabase.NewSupabase(ctx, supClient)
//	logger := logr.DefaultLogger
//	return all.InitiateSQLServices(db, logger)
//}

func getPostgresDatabase(ctx context.Context) (isql.Engine, error) {
	parsePostgresConfig()
	conn, err := sqlib.GetPostgresConnection(TrackerConfig.Database.Postgres)
	if err != nil {
		return nil, err
	}
	go pingPostgresDatabasePeriodically(ctx, TrackerConfig.Database.Postgres, conn, logr.DefaultLogger)

	return postgres.NewPostgres(ctx, conn).ShowSQL(true), nil
}

func pingPostgresDatabasePeriodically(ctx context.Context, cfg sqlib.PostgresConfig, conn *sql.Conn, logger logr.Logger) {
	t5 := time.NewTicker(5 * time.Minute)
	for range t5.C {
		if err := conn.PingContext(ctx); err != nil {
			logger.Errorw("Database connection closed", "error", err.Error())
			conn, err = sqlib.GetPostgresConnection(cfg)
			if err != nil {
				logger.Errorw("couldn't create database connection", "error", err.Error())
			}

			db := postgres.NewPostgres(ctx, conn).ShowSQL(true)
			all.InitiateSQLServices(styx.UnitOfWork{SQL: db}, logger)
			logger.Infow("New connection established")
		}
	}
}

func parsePostgresConfig() {
	user, ok := os.LookupEnv("POSTGRES_USER")
	if ok {
		TrackerConfig.Database.Postgres.User = user
	}
	pass, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if ok {
		TrackerConfig.Database.Postgres.Password = pass
	}
	name, ok := os.LookupEnv("POSTGRES_DB")
	if ok {
		TrackerConfig.Database.Postgres.Name = name
	}
	host, ok := os.LookupEnv("POSTGRES_HOST")
	if ok {
		TrackerConfig.Database.Postgres.Host = host
	}
	port, ok := os.LookupEnv("POSTGRES_PORT")
	if ok {
		TrackerConfig.Database.Postgres.Port = port
	}
	ssl, ok := os.LookupEnv("POSTGRES_SSL_MODE")
	if ok {
		TrackerConfig.Database.Postgres.SSLMode = ssl
	}
}

func syncTables(db isql.Engine) error {
	return db.Sync(
		models.Profile{},
		models.Contacts{},
		models.Wallet{},
		models.Transaction{},
		models.TxnCategory{},
		models.TxnSubcategory{},
		models.Event{},
		models.AICache{},
		models.Budget{},
	)
}

// LoadAICacheIntoMemory loads all persisted AI cache rows into the in-memory cache.
func LoadAICacheIntoMemory() {
	if sqlDB == nil {
		return
	}
	var rows []models.AICache
	if err := sqlDB.Table(models.AICache{}.TableName()).FindMany(&rows); err != nil {
		logr.DefaultLogger.Errorw("Failed to load AI cache", "error", err.Error())
		return
	}
	for _, row := range rows {
		_ = cache.SetCache(row.InputText, row.SubcategoryID, -1)
	}
	logr.DefaultLogger.Infow("AI cache loaded from DB", "count", len(rows))
}

// PingDatabase checks if the database connection is healthy.
func PingDatabase() error {
	if sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := sqlDB.Exec("SELECT 1")
	return err
}

// InsertAICache persists a single AI cache entry to the database.
func InsertAICache(entry models.AICache) error {
	if sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := sqlDB.Table(models.AICache{}.TableName()).InsertOne(entry)
	return err
}
