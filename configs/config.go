package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/masudur-rahman/expense-tracker-bot/modules/cache"

	"github.com/masudur-rahman/styx/sql/postgres/lib"
)

var TrackerConfig ExpenseConfiguration

type ExpenseConfiguration struct {
	Telegram     Telegram            `json:"telegram" yaml:"telegram"`
	Database     DatabaseConfig      `json:"database" yaml:"database"`
	Cache        cache.Config        `json:"cache" yaml:"cache"`
	System       SystemConfig        `json:"system" yaml:"system"`
	WebDashboard WebDashboardConfig  `json:"webDashboard" yaml:"webDashboard"`
}

// WebDashboardConfig holds settings for the optional web dashboard.
type WebDashboardConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	JWTSecret     string `json:"jwtSecret" yaml:"jwtSecret"`
	RefreshSecret string `json:"refreshSecret" yaml:"refreshSecret"`
	BotUsername   string `json:"botUsername" yaml:"botUsername"`
	CORSOrigin   string `json:"corsOrigin" yaml:"corsOrigin"`
	Port          string `json:"port" yaml:"port"`
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

type PDFGenerator string

const (
	PDFGeneratorWkhtmltopdf PDFGenerator = "wkhtmltopdf"
	PDFGeneratorChromeDP    PDFGenerator = "chromedp"
)

type SystemConfig struct {
	PDFGenerator  PDFGenerator `json:"pdfGenerator" yaml:"pdfGenerator"`
	AIGenerator   string       `json:"aiGenerator" yaml:"aiGenerator"`
	GeminiKey     string       `json:"geminiKey" yaml:"geminiKey"`
	OpenRouterKey string       `json:"openRouterKey" yaml:"openRouterKey"`
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

func (c *ExpenseConfiguration) OverrideWithEnv() {
	if token := os.Getenv("EXPENSE_BOT_TOKEN"); token != "" {
		c.Telegram.Secret = token
	}
	if user := os.Getenv("EXPENSE_BOT_USER"); user != "" {
		c.Telegram.User = user
	}

	if dbPass := os.Getenv("EXPENSE_DB_PASS"); dbPass != "" {
		if c.Database.Type == DatabasePostgres {
			c.Database.Postgres.Password = dbPass
		}
	}

	if redisPass := os.Getenv("EXPENSE_REDIS_PASS"); redisPass != "" {
		c.Cache.Redis.Password = redisPass
	}

	// AI Configuration Overrides
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		c.System.GeminiKey = geminiKey
		if c.System.AIGenerator == "" {
			c.System.AIGenerator = "gemini"
		}
	}
	if orKey := os.Getenv("OPENROUTER_API_KEY"); orKey != "" {
		c.System.OpenRouterKey = orKey
		if c.System.AIGenerator == "" {
			c.System.AIGenerator = "open-router"
		}
	}

	// Web Dashboard Overrides
	if os.Getenv("WEB_ENABLED") == "true" {
		c.WebDashboard.Enabled = true
	}
	if secret := os.Getenv("WEB_JWT_SECRET"); secret != "" {
		c.WebDashboard.JWTSecret = secret
	}
	if secret := os.Getenv("WEB_REFRESH_SECRET"); secret != "" {
		c.WebDashboard.RefreshSecret = secret
	}
	if origin := os.Getenv("WEB_CORS_ORIGIN"); origin != "" {
		c.WebDashboard.CORSOrigin = origin
	}
	if username := os.Getenv("WEB_BOT_USERNAME"); username != "" {
		c.WebDashboard.BotUsername = username
	}
	if port := os.Getenv("WEB_PORT"); port != "" {
		c.WebDashboard.Port = port
	}
}
