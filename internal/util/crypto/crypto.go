package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func GenerateSecretKey() (string, error) {
	// Generate a salt for Argon2 key derivation
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Derive a key using Argon2
	secretKey := argon2.IDKey([]byte("password"), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", secretKey), nil
}
