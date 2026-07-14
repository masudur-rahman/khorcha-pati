package validator

import (
	"regexp"
)

var (
	// ShortNameRegex allows alphanumeric characters, dash, and underscore. No whitespace or other special characters.
	ShortNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

	// DisplayNameRegex allows alphanumeric characters, dash, underscore, and spaces, but no leading/trailing spaces.
	DisplayNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]([a-zA-Z0-9\-_ ]*[a-zA-Z0-9\-_])?$`)

	// WalletNameRegex is DisplayNameRegex plus an internal apostrophe (e.g. "Masud's Savings"). No leading/trailing apostrophe.
	WalletNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]([a-zA-Z0-9\-_' ]*[a-zA-Z0-9\-_])?$`)
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

// IsValidWalletName validates a wallet display name, additionally allowing an internal apostrophe.
func IsValidWalletName(s string) bool {
	if s == "" {
		return true // Optional; handled separately
	}
	return WalletNameRegex.MatchString(s)
}
