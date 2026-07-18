package configs

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/masudur-rahman/khorcha-pati/infra/logr"
	"github.com/masudur-rahman/khorcha-pati/models"
	"github.com/masudur-rahman/khorcha-pati/modules/cache"
	"github.com/masudur-rahman/khorcha-pati/modules/google"
	"github.com/masudur-rahman/khorcha-pati/services"
	"github.com/masudur-rahman/khorcha-pati/services/all"

	"github.com/masudur-rahman/styx"
	isql "github.com/masudur-rahman/styx/sql"
	"github.com/masudur-rahman/styx/sql/postgres"
	"github.com/masudur-rahman/styx/sql/sqlite"
	"github.com/masudur-rahman/styx/sql/sqlite/lib"

	_ "github.com/lib/pq"
)

// sqlDB holds a reference to the database engine for utility functions.
var (
	sqlDB isql.Engine
	dbMu  sync.Mutex
)

// GetUnitOfWork returns a UnitOfWork wrapping the active database engine.
func GetUnitOfWork() styx.UnitOfWork {
	return styx.UnitOfWork{SQL: &safeEngine{engine: sqlDB, mu: &dbMu}}
}

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

func getSQLiteDatabase(_ context.Context) (isql.Engine, error) {
	conn, err := lib.GetSQLiteConnection(google.DatabasePath())
	if err != nil {
		return nil, err
	}
	return sqlite.NewSQLite(conn), nil
}

func initializeSQLServices(uow styx.UnitOfWork) error {
	dbMu.Lock()
	sqlDB = uow.SQL
	dbMu.Unlock()

	// Use safeEngine for services to ensure thread safety
	safe := &safeEngine{engine: uow.SQL, mu: &dbMu}
	uow.SQL = safe

	if err := syncTables(uow.SQL); err != nil {
		return err
	}
	if err := fixNullZeroValues(uow.SQL); err != nil {
		return err
	}
	if err := backfillMobileSuffix(uow.SQL); err != nil {
		return err
	}
	all.InitiateSQLServices(uow, logr.DefaultLogger)

	if err := all.GetServices().Access.Seed(buildAccessSeed()); err != nil {
		return fmt.Errorf("seed access control: %w", err)
	}

	return all.GetServices().Txn.UpdateTxnCategories()
}

// buildAccessSeed maps the config's access-control bootstrap into a seed.
// Applied additively every boot; existing rows and settings are never touched.
func buildAccessSeed() services.AccessSeed {
	tg := TrackerConfig.Telegram
	text := "🔒 This is a private test instance of Khorcha-Pati."
	if tg.LiveBotURL != "" {
		text += "\nUse the live bot: " + tg.LiveBotURL
	}
	if tg.LiveDashboardURL != "" {
		text += "\nDashboard: " + tg.LiveDashboardURL
	}
	return services.AccessSeed{
		Restricted:   tg.AllowedUsersOnly,
		AllowedUsers: tg.AllowedUsers,
		ReplyText:    text,
		Owner:        tg.BotOwner,
	}
}

// fixNullZeroValues patches existing rows where styx v1.2.x inserted NULL for zero-value fields.
func fixNullZeroValues(db isql.Engine) error {
	ctx := context.Background()
	stmts := []string{
		`UPDATE "transaction" SET deleted_at = 0 WHERE deleted_at IS NULL`,
		`UPDATE "wallet" SET version = 0 WHERE version IS NULL`,
		`UPDATE "contacts" SET net_balance = 0 WHERE net_balance IS NULL`,
		`UPDATE "contacts" SET last_txn_timestamp = 0 WHERE last_txn_timestamp IS NULL`,
		`UPDATE "budget" SET alert_at = 80 WHERE alert_at IS NULL`,
		`UPDATE "ai_cache" SET intent = '' WHERE intent IS NULL`,
		`UPDATE "refresh_token" SET revoked = 0 WHERE revoked IS NULL`,
		`UPDATE "profile" SET is_admin = false WHERE is_admin IS NULL`,
		`UPDATE "profile" SET is_active = true WHERE is_active IS NULL`,
		`UPDATE "profile" SET created_at = 0 WHERE created_at IS NULL`,
		`UPDATE "wallet" SET created_at = 0 WHERE created_at IS NULL`,
		`UPDATE "contacts" SET created_at = 0 WHERE created_at IS NULL`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("fix null values: %w", err)
		}
	}
	return nil
}

func getPostgresDatabase(ctx context.Context) (isql.Engine, error) {
	parsePostgresConfig()

	pool, err := sql.Open("postgres", TrackerConfig.Database.Postgres.String())
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}

	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)

	if err := pool.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	go pingPostgresDatabasePeriodically(logr.DefaultLogger)

	return postgres.NewPostgres(pool).ShowSQL(true), nil
}

// pingPostgresDatabasePeriodically logs Postgres health. *sql.DB handles reconnections automatically.
func pingPostgresDatabasePeriodically(logger logr.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := PingDatabase(); err != nil {
			logger.Warnw("Postgres health check failed", "error", err.Error())
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
		context.Background(),
		models.Profile{},
		models.Contacts{},
		models.Wallet{},
		models.Transaction{},
		models.TxnCategory{},
		models.TxnSubcategory{},
		models.Event{},
		models.AICache{},
		models.Budget{},
		models.RefreshToken{},
		models.Setting{},
		models.AllowedUser{},
	)
}

// LoadAICacheIntoMemory loads all persisted AI cache rows into the in-memory cache.
func LoadAICacheIntoMemory() {
	if err := PingDatabase(); err != nil {
		return
	}

	ctx := context.Background()
	var rows []models.AICache
	if err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).FindMany(ctx, &rows); err != nil {
		logr.DefaultLogger.Errorw("Failed to load AI cache", "error", err.Error())
		return
	}
	for _, row := range rows {
		setAICacheMemory(row)
	}
	logr.DefaultLogger.Infow("AI cache loaded from DB", "count", len(rows))
}

// PingDatabase checks if the database connection is healthy.
func PingDatabase() error {
	dbMu.Lock()
	defer dbMu.Unlock()
	if sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := sqlDB.Exec(context.Background(), "SELECT 1")
	return err
}

// InsertAICache persists a single AI cache entry to the database.
func InsertAICache(entry models.AICache) error {
	if sqlDB == nil {
		return fmt.Errorf("database not initialized")
	}
	_, err := GetUnitOfWork().SQL.Table(models.AICache{}.TableName()).InsertOne(context.Background(), entry)
	return err
}

// backfillMobileSuffix populates the mobile_suffix lookup key for profiles
// created before the column existed. One-time cost at startup; subsequent
// writes maintain the suffix in the user repo.
func backfillMobileSuffix(db isql.Engine) error {
	ctx := context.Background()
	profiles := make([]models.Profile, 0)
	if err := db.Table(models.Profile{}.TableName()).FindMany(ctx, &profiles); err != nil {
		return fmt.Errorf("backfill mobile suffix: %w", err)
	}
	for i := range profiles {
		suffix := models.PhoneSuffix(profiles[i].MobileNumber)
		if suffix == "" || profiles[i].MobileSuffix == suffix {
			continue
		}
		profiles[i].MobileSuffix = suffix
		if err := db.Table(profiles[i].TableName()).ID(profiles[i].ID).UpdateOne(ctx, &profiles[i]); err != nil {
			return fmt.Errorf("backfill mobile suffix: %w", err)
		}
	}
	return nil
}

// SeedAdminUser marks the configured bot owner as admin.
func SeedAdminUser() {
	username := TrackerConfig.Telegram.BotOwner
	if username == "" {
		return
	}

	ctx := context.Background()
	var profile models.Profile
	found, err := GetUnitOfWork().SQL.Table(profile.TableName()).FindOne(ctx, &profile, models.Profile{Username: username})
	if err != nil || !found {
		return
	}

	if profile.IsAdmin {
		return
	}

	profile.IsAdmin = true
	if err := GetUnitOfWork().SQL.Table(profile.TableName()).ID(profile.ID).MustCols("is_admin").UpdateOne(ctx, profile); err != nil {
		logr.DefaultLogger.Errorw("Failed to seed admin user", "username", username, "error", err.Error())
	}
}

// safeEngine is a thread-safe wrapper around isql.Engine.
type safeEngine struct {
	engine isql.Engine
	mu     *sync.Mutex
}

// Transaction control

func (s *safeEngine) BeginTx(ctx context.Context) (isql.Engine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	engine, err := s.engine.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	return &safeEngine{engine: engine, mu: s.mu}, nil
}

func (s *safeEngine) Commit() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Commit()
}

func (s *safeEngine) Rollback() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Rollback()
}

// Fluent builder methods — no mutex, these build query state only.

func (s *safeEngine) Table(name string) isql.Engine {
	return &safeEngine{engine: s.engine.Table(name), mu: s.mu}
}

func (s *safeEngine) ID(id any) isql.Engine {
	return &safeEngine{engine: s.engine.ID(id), mu: s.mu}
}

func (s *safeEngine) In(col string, values ...any) isql.Engine {
	return &safeEngine{engine: s.engine.In(col, values...), mu: s.mu}
}

func (s *safeEngine) Where(cond string, args ...any) isql.Engine {
	return &safeEngine{engine: s.engine.Where(cond, args...), mu: s.mu}
}

func (s *safeEngine) Columns(cols ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Columns(cols...), mu: s.mu}
}

func (s *safeEngine) AllCols() isql.Engine {
	return &safeEngine{engine: s.engine.AllCols(), mu: s.mu}
}

func (s *safeEngine) MustCols(cols ...string) isql.Engine {
	return &safeEngine{engine: s.engine.MustCols(cols...), mu: s.mu}
}

func (s *safeEngine) MustFilterCols(cols ...string) isql.Engine {
	return &safeEngine{engine: s.engine.MustFilterCols(cols...), mu: s.mu}
}

func (s *safeEngine) ShowSQL(showSQL bool) isql.Engine {
	return &safeEngine{engine: s.engine.ShowSQL(showSQL), mu: s.mu}
}

func (s *safeEngine) OrderBy(col string, direction ...string) isql.Engine {
	return &safeEngine{engine: s.engine.OrderBy(col, direction...), mu: s.mu}
}

func (s *safeEngine) Limit(n int64) isql.Engine {
	return &safeEngine{engine: s.engine.Limit(n), mu: s.mu}
}

func (s *safeEngine) Offset(n int64) isql.Engine {
	return &safeEngine{engine: s.engine.Offset(n), mu: s.mu}
}

func (s *safeEngine) Distinct() isql.Engine {
	return &safeEngine{engine: s.engine.Distinct(), mu: s.mu}
}

func (s *safeEngine) GroupBy(cols ...string) isql.Engine {
	return &safeEngine{engine: s.engine.GroupBy(cols...), mu: s.mu}
}

func (s *safeEngine) Having(cond string, args ...any) isql.Engine {
	return &safeEngine{engine: s.engine.Having(cond, args...), mu: s.mu}
}

func (s *safeEngine) Or(cond string, args ...any) isql.Engine {
	return &safeEngine{engine: s.engine.Or(cond, args...), mu: s.mu}
}

func (s *safeEngine) Like(col, pattern string) isql.Engine {
	return &safeEngine{engine: s.engine.Like(col, pattern), mu: s.mu}
}

func (s *safeEngine) NotLike(col, pattern string) isql.Engine {
	return &safeEngine{engine: s.engine.NotLike(col, pattern), mu: s.mu}
}

func (s *safeEngine) Exists(subquery string, args ...any) isql.Engine {
	return &safeEngine{engine: s.engine.Exists(subquery, args...), mu: s.mu}
}

func (s *safeEngine) NotExists(subquery string, args ...any) isql.Engine {
	return &safeEngine{engine: s.engine.NotExists(subquery, args...), mu: s.mu}
}

func (s *safeEngine) Count(col string, alias ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Count(col, alias...), mu: s.mu}
}

func (s *safeEngine) Sum(col string, alias ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Sum(col, alias...), mu: s.mu}
}

func (s *safeEngine) Avg(col string, alias ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Avg(col, alias...), mu: s.mu}
}

func (s *safeEngine) Min(col string, alias ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Min(col, alias...), mu: s.mu}
}

func (s *safeEngine) Max(col string, alias ...string) isql.Engine {
	return &safeEngine{engine: s.engine.Max(col, alias...), mu: s.mu}
}

func (s *safeEngine) Paginate(page, perPage int64) isql.Engine {
	return &safeEngine{engine: s.engine.Paginate(page, perPage), mu: s.mu}
}

func (s *safeEngine) Join(table, condition string) isql.Engine {
	return &safeEngine{engine: s.engine.Join(table, condition), mu: s.mu}
}

func (s *safeEngine) LeftJoin(table, condition string) isql.Engine {
	return &safeEngine{engine: s.engine.LeftJoin(table, condition), mu: s.mu}
}

func (s *safeEngine) RightJoin(table, condition string) isql.Engine {
	return &safeEngine{engine: s.engine.RightJoin(table, condition), mu: s.mu}
}

func (s *safeEngine) InnerJoin(table, condition string) isql.Engine {
	return &safeEngine{engine: s.engine.InnerJoin(table, condition), mu: s.mu}
}

func (s *safeEngine) WithDeleted() isql.Engine {
	return &safeEngine{engine: s.engine.WithDeleted(), mu: s.mu}
}

func (s *safeEngine) EnableValidation(enable bool) isql.Engine {
	return &safeEngine{engine: s.engine.EnableValidation(enable), mu: s.mu}
}

// Execution methods — mutex required.

func (s *safeEngine) FindOne(ctx context.Context, document any, filter ...any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.FindOne(ctx, document, filter...)
}

func (s *safeEngine) FindMany(ctx context.Context, documents any, filter ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.FindMany(ctx, documents, filter...)
}

func (s *safeEngine) InsertOne(ctx context.Context, document any) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.InsertOne(ctx, document)
}

func (s *safeEngine) InsertMany(ctx context.Context, documents []any) ([]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.InsertMany(ctx, documents)
}

func (s *safeEngine) UpdateOne(ctx context.Context, document any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.UpdateOne(ctx, document)
}

func (s *safeEngine) DeleteOne(ctx context.Context, filter ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.DeleteOne(ctx, filter...)
}

func (s *safeEngine) ForceDelete(ctx context.Context, filter ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.ForceDelete(ctx, filter...)
}

func (s *safeEngine) Restore(ctx context.Context, filter ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Restore(ctx, filter...)
}

func (s *safeEngine) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Query(ctx, query, args...)
}

func (s *safeEngine) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Exec(ctx, query, args...)
}

func (s *safeEngine) Sync(ctx context.Context, tables ...any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Sync(ctx, tables...)
}

func (s *safeEngine) DropTable(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.DropTable(ctx, name)
}

func (s *safeEngine) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine.Close()
}
