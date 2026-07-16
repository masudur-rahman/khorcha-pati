package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePhoneNumber(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "8801712345678", NormalizePhoneNumber("+880 1712-345678"))
	assert.Equal(t, "01712345678", NormalizePhoneNumber("01712345678"))
	assert.Equal(t, "", NormalizePhoneNumber("abc"))
}

// All country-code variants of a number must produce the same lookup key.
func TestPhoneSuffix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with plus and country code", "+8801712345678", "12345678"},
		{"country code no plus", "8801712345678", "12345678"},
		{"local with leading zero", "01712345678", "12345678"},
		{"international 00 prefix", "008801712345678", "12345678"},
		{"formatted", "+880 1712-345678", "12345678"},
		{"8-digit subscriber with country code", "+6581234567", "81234567"},
		{"8-digit subscriber local", "81234567", "81234567"},
		{"shorter than 8 digits kept as-is", "12345", "12345"},
		{"no digits", "nobody", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, PhoneSuffix(tc.input))
		})
	}
}

// Verification: same-suffix numbers from different countries must not match.
func TestPhoneNumbersMatch(t *testing.T) {
	t.Parallel()
	assert.True(t, PhoneNumbersMatch("+8801712345678", "01712345678"))
	assert.True(t, PhoneNumbersMatch("01712345678", "8801712345678"))
	assert.True(t, PhoneNumbersMatch("+6581234567", "81234567"))
	assert.False(t, PhoneNumbersMatch("+8801712345678", "+6512345678"))
	assert.False(t, PhoneNumbersMatch("", "01712345678"))
}
