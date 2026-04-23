package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const otpMax = 1000000 // 6-digit range: 000000–999999

// GenerateOTP returns a cryptographically random 6-digit string.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(otpMax))
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
