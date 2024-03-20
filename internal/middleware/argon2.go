package middleware

/*
* Argon2 hashing algorithm implementation
* The argon2.go file contains the Argon2 hashing algorithm implementation.
* The Argon2 hashing algorithm is a password-hashing function that summarizes the state of the art in the design of memory-hard functions and can be used to hash passwords for credential storage, key derivation, or other applications.
* The Argon2 algorithm was the winner of the Password Hashing Competition in 2015.
* The Argon2 algorithm has three primary variants: Argon2d, Argon2i, and Argon2id.
* The Argon2id variant is recommended for password hashing and password-based key derivation.
* The Argon2 algorithm has three primary parameters:
* 1. Memory: The amount of memory used by the algorithm.
* 2. Iterations: The number of iterations of the algorithm.
* 3. Parallelism: The number of threads used by the algorithm.
* The Argon2 algorithm also uses a salt and a key length.
* The Argon2 algorithm is implemented using the golang.org/x/crypto/argon2 package.
* The Argon2 encoded hash structure is as follows:
* $argon2id$v=19$m=65536,t=3,p=2$1DHhXs0CPbS5lZrFDBHklA$FFtdaqsk3uSk5fJn2CDrXbSyHGO65352pllKZxCx/BQ
* Where the fields are as follows:
* 1. argon2id: The Argon2 variant.
* 2. v=19: The Argon2 version.
* 3. m=65536: The amount of memory used by the algorithm.
* 4. t=3: The number of iterations of the algorithm.
* 5. p=2: The number of threads used by the algorithm.
* 6. 1DHhXs0CPbS5lZrFDBHklA: The salt.
* 7. FFtdaqsk3uSk5fJn2CDrXbSyHGO65352pllKZxCx/BQ: The hashed password.
 */

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type Argon2 struct {
	hash        []byte
	Salt        []byte
	encodedHash string
	Params      params
	GetParams   func() *params
}

type Argon2MapParams map[string]interface{}

// 	GetMemory() uint32
// 	GetIterations() uint32
// 	GetParallelism() uint8
// 	GetSaltLength() uint32
// 	GetKeyLength() uint32
// }

type Argon2Params interface {
	GetMemory() uint32
	GetIterations() uint32
	GetParallelism() uint8
	GetSaltLength() uint32
	GetKeyLength() uint32
}

func DefaultParams() *params {
	return &params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32, // AES256 compatible key length
	}
}

// HashesMatch compares a password and plaintext hash to validate the password
func HashesMatch(password string, salt []byte) (bool, error) {
	encodedHash := NewArgon2().GetPrintableKeyWithSalt(salt)

	return ComparePasswordAndHash(password, encodedHash)
}

// InitArgon initializes the Argon2 hashing algorithm and returns an Argon2 struct
// Note that this function genereates a new salt and hash
// May be in conjunction with the NewArgon2 function
func (a *Argon2) InitArgon(password string) Argon2 {

	p := DefaultParams()

	salt, err := generateRandomBytes(p.saltLength)

	fmt.Println("Generated salt: ", salt)

	if err != nil {
		fmt.Println("Error generating random bytes for salt")
		panic(err)
	}

	hash, encodedHash, err := generateFromPassword(password, p)

	fmt.Println("Generated hash: ", hash)
	fmt.Println("Generated encoded hash: ", encodedHash)

	if err != nil {
		panic(err)
	}

	// // Base64 encode the salt and hashed password.
	// b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	// b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	getParams := func() *params {
		return p
	}

	a = &Argon2{
		hash:        hash,
		Params:      *p,
		encodedHash: encodedHash,
		Salt:        salt,
		GetParams:   getParams,
	}

	match, err := ComparePasswordAndHash(password, a.encodedHash)

	if err != nil {
		fmt.Println("Error comparing password and hash")
		panic(err)
	}

	if match {
		fmt.Printf("\033[0;32m%s\033[0m\n", "Password and hash match")
	} else {
		fmt.Printf("\033[0;31m%s\033[0m\n", "Password and hash do not match")
	}

	return *a
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

func (a *Argon2) InitArgonWithSalt(password string, salt string) Argon2 {
	p := DefaultParams()

	hash, err := generateFromPasswordWithSalt(password, salt, p)
	a.encodedHash = HashToEncodedHash(p, hash, []byte(salt))

	if err != nil {
		panic(err)
	}

	a.hash = hash
	a.Params = *p
	a.Salt = []byte(salt)

	return *a
}

func (a *Argon2) GetPrintableKey() string {
	p := a.Params
	salt, err := generateRandomBytes(p.saltLength)

	if err != nil {
		fmt.Println("Error generating random bytes for salt")
		panic(err)
	}

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(a.hash)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash
}

func (a *Argon2) GetPrintableKeyWithSalt(salt []byte) string {
	// p := GetParams()
	p := a.Params

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(a.Salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(a.hash) // TODO: Need to remove - we don't store the hash now

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash
}

func (a *Argon2) GetHash() []byte {
	if a.hash == nil {
		panic("Hash not generated")
	}

	return a.hash
}

func NewArgon2() *Argon2 {
	return &Argon2{}
}

func generateFromPassword(password string, p *params) (hash []byte, encodedHash string, err error) {
	fmt.Printf("Generating argon2 hash with password \033[1;36m%s\033[0m and salt length: %v\n", password, p.saltLength)

	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return nil, "", err
	}

	// Pass the plaintext password, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the password using the Argon2id
	// variant.
	hash = argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return a string using the standard encoded hash representation.
	encodedHash = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return hash, encodedHash, nil
}

func generateFromPasswordWithSalt(password string, salt string, p *params) (hash []byte, err error) {
	hash = argon2.IDKey([]byte(password), []byte(salt), p.iterations, p.memory, p.parallelism, p.keyLength)
	return hash, nil
}

func (a *Argon2) GetEncodedHash() string {
	return a.encodedHash
}

// HashToEncodedHash hashes a password and salt and returns the encoded hash.
// Ex:
// hash: []byte("FFtdaqsk3uSk5fJn2CDrXbSyHGO65352pllKZxCx/BQ")
// salt: []byte("salt")
func HashToEncodedHash(p *params, hash []byte, salt []byte) string {
	if p == nil {
		p = DefaultParams()
	}
	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)
}

// GenerateRandomBytes generates a random byte slice of length n.
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b) // Crypto secure random number generator
	if err != nil {
		return nil, err
	}

	return b, nil
}

// ComparePasswordAndHash compares a password and encoded hash to check if the
// password matches the hash. Only the encoded hash and the plaintext password are needed.
func ComparePasswordAndHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := DecodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the password using the same parameters.
	pwHash := argon2.IDKey(
		[]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, pwHash) == 1 {
		return true, nil
	}
	return false, nil
}

// Decode hash decodes an encoded hash string and returns the parameters, salt, and hash (byte array).
func DecodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	fmt.Println("Decoding hash " + "\033[0;35m" + encodedHash + "\033[0m")
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func Base64ToBytes(encoded string) ([]byte, error) {
	return base64.RawStdEncoding.Strict().DecodeString(encoded)
}
