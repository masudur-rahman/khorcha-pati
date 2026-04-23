package auth

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOTP_format(t *testing.T) {
	t.Parallel()
	otp, err := GenerateOTP()
	require.NoError(t, err)
	assert.Len(t, otp, 6)
	for _, c := range otp {
		assert.True(t, unicode.IsDigit(c), "expected digit, got %c", c)
	}
}

func TestGenerateOTP_unique(t *testing.T) {
	t.Parallel()
	seen := make(map[string]bool)
	for range 20 {
		otp, err := GenerateOTP()
		require.NoError(t, err)
		seen[otp] = true
	}
	assert.Greater(t, len(seen), 1, "expected multiple unique OTPs")
}
