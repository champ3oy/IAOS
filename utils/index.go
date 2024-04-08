package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomCode generates a random code of specified length
func GenerateRandomCode(length int) (string, error) {
	// Calculate the number of bytes needed for the specified length
	numBytes := length / 2
	if length%2 != 0 {
		numBytes++
	}

	// Generate random bytes
	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Convert random bytes to hexadecimal string
	code := hex.EncodeToString(randomBytes)[:length]
	return code, nil
}
