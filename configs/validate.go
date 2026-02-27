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
	"PARSE_APP_ID",
	"PARSE_REST_API_KEY",
	"PARSE_SERVER_URL",
}

// Validate checks that all required environment variables are present and
// non-empty. Call this at the very start of main() before initialising
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
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf(
		"missing required environment variables: %s\n"+
			"Copy .env.example to .env and fill in the values, "+
			"or set them as environment variables.",
		strings.Join(missing, ", "),
	)
}
