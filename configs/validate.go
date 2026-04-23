// Package configs provides configuration loading and validation.
package configs

import (
	"fmt"
	"os"
	"strings"
)

// requiredEnvVars lists environment variables that MUST be set for the bot
// to start. Add new required keys here as they are introduced.
var requiredEnvVars = []string{
	"TELEGRAM_BOT_TOKEN",
}

// webRequiredEnvVars are required only when WEB_ENABLED=true.
var webRequiredEnvVars = []string{
	"WEB_JWT_SECRET",
	"WEB_REFRESH_SECRET",
}

// Validate checks that all required environment variables are present and
// non-empty. Call this at the very start of main() before initializing
// anything else so operators get a clear diagnostic on misconfiguration.
//
//	if err := configs.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func Validate() error {
	var missing []string
	for _, key := range requiredEnvVars {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}
	if os.Getenv("WEB_ENABLED") == "true" {
		for _, key := range webRequiredEnvVars {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				missing = append(missing, key)
			}
		}
	}

	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"missing required environment variables: %s",
		strings.Join(missing, ", "),
	)
}
