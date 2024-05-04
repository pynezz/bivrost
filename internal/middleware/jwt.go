package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/cryptoutils"
)

const (
	exp = 9         // Expires in 9 hours
	iss = "bivrost" // Issuer
)

func GetSecretKey() string {
	return cryptoutils.GetBivrostJWTSecret()
}

func VerifyJWTToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		signing := token.Header["alg"]
		fmt.Printf("%v\n", token)
		fmt.Println("Signing method: ", signing)

		exp := token.Claims.(jwt.MapClaims)["exp"].(float64)
		debugExp := fmt.Sprintf("%f", exp)
		util.PrintDebug(fmt.Sprintf("Token expires at: %s", debugExp))

		timeNow := time.Now().Unix()
		fmt.Println("Time now: ", timeNow)

		if exp < float64(timeNow) {
			fmt.Println("Token has expired")
			return nil, fiber.ErrUnauthorized
		}

		// Ensure token's signing method matches
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}

		// Return the secret key to the jwt.Parse function
		return []byte(GetSecretKey()), nil
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error getting claims")
		return nil, err
	}

	sub := fmt.Sprintf("%f", claims["sub"].(float64))
	util.PrintDebug("Subject(user): " + sub)
	aud := fmt.Sprintf("%s", claims["aud"])
	util.PrintDebug("Audience(role): " + aud)

	return token, err
}

func GenerateJWTToken(user User, loginTime time.Time) string {
	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		//  // NB! This might not work. It's a time.Time object
		"exp": loginTime.Add(time.Duration(exp) * time.Hour).Unix(),

		"iss": iss,
		"aud": user.Role,
		"sub": user.UserID,
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(GetSecretKey()))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(tokenString)
	return tokenString
}
