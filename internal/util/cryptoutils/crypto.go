package cryptoutils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/argon2"
)

// GenerateSecretKey generates a secret key using Argon2
func GenerateSecretKey(password string) (string, error) {
	// Generate a salt for Argon2 key derivation
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Derive a key using Argon2
	secretKey := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
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
		// The caller should format this with util.ErrorF
		return 0, err
	}

	return randomID.Uint64(), nil
}

// Get Bivrost JWT secret Argon2 encoded hash from TheValve
func GetBivrostJWTSecret() string {
	// extract "$argon2id$v=19$m=16,t=2,p=1$aGVORDFaUmJrVTgzdm1wQw$Sin4nQSD3lnO7G8XD6lb7Q"
	// decodeArgon2Hash
	// return string(vault.GetVault().GetSection("bivrost").GetEntry("jwt").GetContent("secret")
	return "password"
}

// func DecodeArgon2Hash(encodedHash string) (p *Argon2Params, salt, hash []byte, err error) {
// 	fmt.Println("Decoding hash " + "\033[0;35m" + encodedHash + "\033[0m")
// 	vals := strings.Split(encodedHash, "$")
// 	if len(vals) != 6 {
// 		return nil, nil, nil, errors.New("invalid hash")
// 	}

// 	var version int
// 	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	if version != argon2.Version {
// 		return nil, nil, nil, errors.New("incompatible version")
// 	}

// 	p = &Argon2Params{}
// 	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}

// 	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	p.saltLength = uint32(len(salt))

// 	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	p.keyLength = uint32(len(hash))

// 	return p, salt, hash, nil
// }

// func encodeArgon2Hash(p *Argon2Params, salt, hash []byte) string {
// 	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
// 		argon2.Version, p.memory, p.iterations, p.parallelism,
// 		base64.RawStdEncoding.Strict().EncodeToString(salt),
// 		base64.RawStdEncoding.Strict().EncodeToString(hash))
// }

// // $argon2id$v=19$m=65536,t=3,p=2$1DHhXs0CPbS5lZrFDBHklA$FFtdaqsk3uSk5fJn2CDrXbSyHGO65352pllKZxCx/BQ
// func (a *Argon2) EncodeHash() {
// 	a.EncodedHash = encodeArgon2Hash(&a.params, a.salt, a.hash)
// }

// func (a *Argon2) DecodeHash() error {
// 	var err error
// 	var p *Argon2Params                                      // Declare a variable of type *Argon2Params
// 	p, a.salt, a.hash, err = DecodeArgon2Hash(a.EncodedHash) // Assign the return value to p
// 	if err != nil {
// 		return err
// 	}
// 	a.params = *p // Assign the value of p to a.params
// 	return nil
// }

// func (a *Argon2) ComparePassword(password string) (match bool, err error) {
// 	// Derive the key from the password using the same parameters.
// 	pwHash := argon2.IDKey([]byte(password), a.salt, a.params.iterations, a.params.memory, a.params.parallelism, a.params.keyLength)

// 	// Check that the contents of the hashed passwords are identical. Note
// 	// that we are using the subtle.ConstantTimeCompare() function for this
// 	// to help prevent timing attacks.
// 	if subtle.ConstantTimeCompare(a.hash, pwHash) == 1 {
// 		return true, nil
// 	}
// 	return false, nil
// }

// func InitArgon2(argon2 *Argon2) {
// 	argon2.params.memory = 64 * 1024
// 	argon2.params.iterations = 3
// 	argon2.params.parallelism = 2
// }
