package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomCode(length int) (string, error) {
	numBytes := length / 2
	if length%2 != 0 {
		numBytes++
	}

	randomBytes := make([]byte, numBytes)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	code := hex.EncodeToString(randomBytes)[:length]
	return code, nil
}
