package validator

import (
	"regexp"
)

var (
	// ShortNameRegex allows alphanumeric characters, dash, and underscore. No whitespace or other special characters.
	ShortNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

	// DisplayNameRegex allows alphanumeric characters, dash, underscore, and spaces, but no leading/trailing spaces.
	DisplayNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]([a-zA-Z0-9\-_ ]*[a-zA-Z0-9\-_])?$`)
)

func IsValidShortName(s string) bool {
	return ShortNameRegex.MatchString(s)
}

func IsValidDisplayName(s string) bool {
	if s == "" {
		return true // Usually optional or handled separately
	}
	return DisplayNameRegex.MatchString(s)
}
