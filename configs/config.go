package configs

import (
	"fmt"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"

	"github.com/masudur-rahman/styx/sql/postgres/lib"
)

var TrackerConfig ExpenseConfiguration

type ExpenseConfiguration struct {
	Telegram Telegram       `json:"telegram" yaml:"telegram"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Cache    cache.Config   `json:"cache" yaml:"cache"`
	System   SystemConfig   `json:"system" yaml:"system"`
}

type Telegram struct {
	User   string `json:"user" yaml:"user"`
	Secret string `json:"secret" yaml:"secret"`
}

type DatabaseConfig struct {
	Type DatabaseType `json:"type" yaml:"type"`

	//ArangoDB DBConfigArangoDB `json:"arangodb" yaml:"arangodb"`
	Postgres lib.PostgresConfig `json:"postgres" yaml:"postgres"`
	SQLite   DBConfigSQLite     `json:"sqlite" yaml:"sqlite"`
}

type SystemConfig struct {
	PDFConverter string `json:"pdfConverter" yaml:"pdfConverter"`
}

type DatabaseType string

const (
	DatabaseArangoDB DatabaseType = "arangodb"
	DatabasePostgres DatabaseType = "postgres"
	DatabaseSQLite   DatabaseType = "sqlite"
	DatabaseSupabase DatabaseType = "supabase"
)

type DBConfigArangoDB struct {
	Name     string `json:"name" yaml:"name"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
}

type DBConfigPostgres struct {
	Name     string `json:"name" yaml:"name"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
}

type DBConfigSQLite struct {
	SyncToDrive          bool          `json:"syncToDrive" yaml:"syncToDrive"`
	DisableSyncFromDrive bool          `json:"disableSyncFromDrive" yaml:"disableSyncFromDrive"`
	SyncInterval         time.Duration `json:"syncInterval" yaml:"syncInterval"`
}

func (cp DBConfigPostgres) String() string {
	return fmt.Sprintf("user=%v password=%v dbname=%v host=%v port=%v sslmode=%v", cp.User, cp.Password, cp.Name, cp.Host, cp.Port, cp.SSLMode)
}
