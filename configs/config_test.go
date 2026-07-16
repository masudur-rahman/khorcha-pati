package configs

import (
	"os"
	"testing"

	"github.com/masudur-rahman/go-oneliners"

	"github.com/stretchr/testify/assert"
)

func TestOverrideWithEnv(t *testing.T) {
	// Setup
	os.Setenv("EXPENSE_BOT_TOKEN", "test-token")
	os.Setenv("EXPENSE_BOT_OWNER", "test-owner")
	os.Setenv("EXPENSE_DB_PASS", "test-db-pass")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")

	defer func() {
		os.Unsetenv("EXPENSE_BOT_TOKEN")
		os.Unsetenv("EXPENSE_BOT_OWNER")
		os.Unsetenv("EXPENSE_DB_PASS")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg := &ExpenseConfiguration{}
	cfg.Database.Type = DatabasePostgres // Trigger the nested pass override

	// Act
	cfg.OverrideWithEnv()
	oneliners.PrettyJson(cfg, "Config")

	// Assert
	assert.Equal(t, "test-token", cfg.Telegram.Secret)
	assert.Equal(t, "test-owner", cfg.Telegram.BotOwner)
	assert.Equal(t, "test-db-pass", cfg.Database.Postgres.Password)
	assert.Equal(t, "test-gemini-key", cfg.System.GeminiKey)
	assert.Equal(t, "gemini", cfg.System.AIClassifier)
}

func TestOverrideWithEnv_BotOwnerFallbacks(t *testing.T) {
	t.Run("legacy yaml user key", func(t *testing.T) {
		cfg := &ExpenseConfiguration{}
		cfg.Telegram.User = "legacy-user"

		cfg.OverrideWithEnv()

		assert.Equal(t, "legacy-user", cfg.Telegram.BotOwner)
		assert.Empty(t, cfg.Telegram.User)
	})

	t.Run("legacy env var", func(t *testing.T) {
		os.Setenv("EXPENSE_BOT_USER", "legacy-env-user")
		defer os.Unsetenv("EXPENSE_BOT_USER")

		cfg := &ExpenseConfiguration{}
		cfg.OverrideWithEnv()

		assert.Equal(t, "legacy-env-user", cfg.Telegram.BotOwner)
	})

	t.Run("new env wins over legacy", func(t *testing.T) {
		os.Setenv("EXPENSE_BOT_OWNER", "new-owner")
		os.Setenv("EXPENSE_BOT_USER", "legacy-env-user")
		defer func() {
			os.Unsetenv("EXPENSE_BOT_OWNER")
			os.Unsetenv("EXPENSE_BOT_USER")
		}()

		cfg := &ExpenseConfiguration{}
		cfg.Telegram.User = "legacy-user"
		cfg.OverrideWithEnv()

		assert.Equal(t, "new-owner", cfg.Telegram.BotOwner)
	})
}

func TestIsBotOwner(t *testing.T) {
	cfg := ExpenseConfiguration{}
	cfg.Telegram.BotOwner = "Admin"

	assert.True(t, cfg.IsBotOwner("admin"))
	assert.True(t, cfg.IsBotOwner("Admin"))
	assert.False(t, cfg.IsBotOwner("someone"))
	assert.False(t, ExpenseConfiguration{}.IsBotOwner(""))
}
