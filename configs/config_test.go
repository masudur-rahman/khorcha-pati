package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverrideWithEnv(t *testing.T) {
	// Setup
	os.Setenv("EXPENSE_BOT_TOKEN", "test-token")
	os.Setenv("EXPENSE_BOT_USER", "test-user")
	os.Setenv("EXPENSE_DB_PASS", "test-db-pass")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")

	defer func() {
		os.Unsetenv("EXPENSE_BOT_TOKEN")
		os.Unsetenv("EXPENSE_BOT_USER")
		os.Unsetenv("EXPENSE_DB_PASS")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	cfg := &ExpenseConfiguration{}
	cfg.Database.Type = DatabasePostgres // Trigger the nested pass override

	// Act
	cfg.OverrideWithEnv()

	// Assert
	assert.Equal(t, "test-token", cfg.Telegram.Secret)
	assert.Equal(t, "test-user", cfg.Telegram.User)
	assert.Equal(t, "test-db-pass", cfg.Database.Postgres.Password)
	assert.Equal(t, "test-gemini-key", cfg.System.GeminiKey)
	assert.Equal(t, "gemini", cfg.System.AIGenerator)
}
