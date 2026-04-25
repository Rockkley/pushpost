package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const digits = 6

func Generate() (string, error) {
	maximum := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, maximum)

	if err != nil {
		return "", fmt.Errorf("generate random OTP: %w", err)
	}

	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}
