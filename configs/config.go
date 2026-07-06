package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/masudur-rahman/khorcha-pati/modules/cache"

	"github.com/masudur-rahman/styx/sql/postgres/lib"
)

var TrackerConfig ExpenseConfiguration

type ExpenseConfiguration struct {
	Telegram Telegram       `json:"telegram" yaml:"telegram"`
	Database DatabaseConfig `json:"database" yaml:"database"`
	Cache    cache.Config   `json:"cache" yaml:"cache"`
	System   SystemConfig   `json:"system" yaml:"system"`
	Server   ServerConfig   `json:"server" yaml:"server"`
}

// ServerConfig holds settings for the optional REST API server.
type ServerConfig struct {
	DashboardEnabled bool   `json:"dashboardEnabled" yaml:"dashboardEnabled"`
	JWTSecret        string `json:"jwtSecret" yaml:"jwtSecret"`
	RefreshSecret    string `json:"refreshSecret" yaml:"refreshSecret"`
	BotUsername      string `json:"botUsername" yaml:"botUsername"`
	CORSOrigin       string `json:"corsOrigin" yaml:"corsOrigin"`
	Host             string `json:"host" yaml:"host"`
	Port             int    `json:"port" yaml:"port"`
	BaseURL          string `json:"baseURL" yaml:"baseURL"`
	DashboardURL     string `json:"dashboardURL" yaml:"dashboardURL"`
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
	PDFGenerator PDFGenerator `json:"pdfGenerator" yaml:"pdfGenerator"`
	// AIClassifier selects the classification provider: "gemini", "open-router", or "pool"
	// (sticky rotation + failover across all configured providers).
	AIClassifier string `json:"aiClassifier" yaml:"aiClassifier"`
	// AIStickyWindow is how many requests a provider serves before the pool rotates (0 = default).
	AIStickyWindow int    `json:"aiStickyWindow" yaml:"aiStickyWindow"`
	GeminiKey      string `json:"geminiKey" yaml:"geminiKey"`
	OpenRouterKey  string `json:"openRouterKey" yaml:"openRouterKey"`
	// GeminiModel / OpenRouterModel override the per-provider model. Empty falls back to the
	// package default (Gemini31FlashLite / NVDIANemotron30bFree).
	GeminiModel     string `json:"geminiModel" yaml:"geminiModel"`
	OpenRouterModel string `json:"openRouterModel" yaml:"openRouterModel"`
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
	geminiKey, hasGemini := os.LookupEnv("GEMINI_API_KEY")
	orKey, hasOpenRouter := os.LookupEnv("OPENROUTER_API_KEY")

	if hasGemini {
		c.System.GeminiKey = geminiKey
	}
	if hasOpenRouter {
		c.System.OpenRouterKey = orKey
	}

	switch {
	case hasGemini && hasOpenRouter:
		if c.System.AIClassifier == "" {
			c.System.AIClassifier = "pool" // rotate + failover across both providers
		}
	case hasGemini:
		c.System.AIClassifier = "gemini"
	case hasOpenRouter:
		c.System.AIClassifier = "open-router"
	default:
		c.System.AIClassifier = ""
	}

	if w := os.Getenv("AI_STICKY_WINDOW"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			c.System.AIStickyWindow = n
		}
	}

	if model := os.Getenv("GEMINI_MODEL"); model != "" {
		c.System.GeminiModel = model
	}
	if model := os.Getenv("OPENROUTER_MODEL"); model != "" {
		c.System.OpenRouterModel = model
	}

	// Server Overrides
	if os.Getenv("SERVER_ENABLED") == "true" {
		c.Server.DashboardEnabled = true
	}
	if secret := os.Getenv("SERVER_JWT_SECRET"); secret != "" {
		c.Server.JWTSecret = secret
	}
	if secret := os.Getenv("SERVER_REFRESH_SECRET"); secret != "" {
		c.Server.RefreshSecret = secret
	}
	if origin := os.Getenv("SERVER_CORS_ORIGIN"); origin != "" {
		c.Server.CORSOrigin = origin
	}
	if username := os.Getenv("SERVER_BOT_USERNAME"); username != "" {
		c.Server.BotUsername = username
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}
	if c.Server.Port == 0 {
		c.Server.Port = 6336
	}
	if baseURL := os.Getenv("SERVER_BASE_URL"); baseURL != "" {
		c.Server.BaseURL = baseURL
	}
	if dashboardURL := os.Getenv("SERVER_DASHBOARD_URL"); dashboardURL != "" {
		c.Server.DashboardURL = dashboardURL
	}
}
