package cryptoutils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/pynezz/bivrost/internal/util"
	"golang.org/x/crypto/argon2"
)

// GenerateSecretKey generates a secret key using Argon2
func GenerateSecretKey() (string, error) {
	// Generate a salt for Argon2 key derivation
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Derive a key using Argon2
	// TODO: Don't hardcode the params.
	secretKey := argon2.IDKey([]byte("password"), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", secretKey), nil
}

// GenerateRandomString generates a random string of a given length
func GenerateRandomString(length int) (string, error) {
	// Generate a cryptographically secure random 32-bit integer
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", randomBytes), nil
}

// Generate a cryptographically secure random 32-bit integer within a range
func GenerateRandomInt(min, max int) (uint64, error) {
	// Generate a cryptographically secure random 32-bit integer
	// var randomID big.Int
	randomID, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)-1)) // -1 because the max value is inclusive
	if err != nil {
		util.PrintError("Error generating random ID")
		return 0, err
	}

	return randomID.Uint64(), nil
}
