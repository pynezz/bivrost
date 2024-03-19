package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pynezz/bivrost/internal/util/cryptoutils"
)

const (
	exp = 9         // Expires in 9 hours
	iss = "bivrost" // Issuer
	aud = "bivrost" // Audience
)

func GetSecretKey() string {
	return cryptoutils.GetBivrostJWTSecret()
}

func VerifyJWTToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		signing := token.Header["alg"]
		fmt.Printf("%v\n", token)
		fmt.Println("Signing method: ", signing)

		// Ensure token's signing method matches
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}

		// Return the secret key to the jwt.Parse function
		return []byte(GetSecretKey()), nil
	})

	if err != nil {
		fmt.Println(err)
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error getting claims")
	}

	fmt.Println(claims["sub"])

	return token, err
}

func GenerateJWTToken() {
	// Create a new token object, specifying signing method and the claims
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	// Set token claims
	claims["sub"] = "bivrost"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(GetSecretKey()))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(tokenString)
}
