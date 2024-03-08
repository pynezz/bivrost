package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pynezz/bivrost/internal/util"
)

// Claims defines the structure of the JWT claims.
type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// AuthMiddleware is a middleware that checks if the user is authenticated.
func AuthMiddleware(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("Authenticating user...")
		// Extract the token from the Authorization header.
		authHeader := c.Get("Authorization")
		splits := strings.Split(authHeader, " ")
		if len(splits) != 2 || splits[0] != "Bearer" {
			fmt.Printf("Splits: %d\n Splits[0]:%s \n", len(splits), splits[0])
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}
		tokenString := splits[1]

		// ---

		// tokenString := c.Get("Authorization")
		fmt.Println("Token: ", tokenString)

		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			signing := token.Header["alg"]
			fmt.Printf("%v\n", token)
			fmt.Println("Signing method: ", signing)
			// Ensure token's signing method matches
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}

			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
		}

		util.PrintColorAndBgBold(util.Green, util.Cyan, "[+] User is authenticated 🎉")

		// Token is valid, proceed with the request
		return c.Next()
	}
}

// GenerateToken generates a JWT token for a user.
func GenerateToken(userID, secretKey string) (string, error) {
	// Define token claims
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(9 * time.Hour)), // Token expires in 9 hours (one work-day + one hour overtime 😄)
			Issuer:    "Bachelorprosjekt",
			Subject:   "UserAuthentication",
		},
	}

	// Create a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
