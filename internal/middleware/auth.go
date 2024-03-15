package middleware // Could maybe rename to handlers

import (
	"encoding/base64"
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

// CREATE TABLE auth_methods (
//     AuthMethodID INTEGER PRIMARY KEY,
//     Description TEXT
// );

// /* Populate auth_methods with initial data */
// INSERT INTO auth_methods (AuthMethodID, Description)
// VALUES
//     (1, 'Password'),
//     (2, 'WebAuthn');

// CREATE TABLE user_sessions ( -- This table will store the user sessions
//     SessionID TEXT PRIMARY KEY,
//     UserID INTEGER NOT NULL,
//     Token TEXT NOT NULL, /* The user session is a JWT Token */
//     FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
// );

// CREATE TABLE webauthn_auth (    -- This table will store essential data for WebAuthn authentication
//     CredentialID TEXT PRIMARY KEY,
//     UserID INTEGER NOT NULL,
//     PublicKey TEXT NOT NULL,
//     UserHandle TEXT NOT NULL,
//     SignatureCounter INTEGER NOT NULL,
//     CreatedAt TEXT DEFAULT (datetime('now')),
//     FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
// );

// CREATE TABLE password_auth ( -- This table will store the related rows for password authentication
//     UserID INTEGER PRIMARY KEY,
//     Enabled BOOLEAN DEFAULT 1, -- SQLite uses 1 for TRUE
//     PasswordHash TEXT NOT NULL, -- Argon2 hash
//     FOREIGN KEY (UserID) REFERENCES users(UserID) ON DELETE CASCADE
// );

type AuthMethod struct {
	AuthMethodID int
	Description  string
}

type UserSession struct {
	SessionID string
	UserID    int
	Token     string
}

type WebAuthnAuth struct {
	CredentialID     string
	UserID           int
	PublicKey        string
	UserHandle       string
	SignatureCounter int
	CreatedAt        time.Time
}

type PasswordAuth struct {
	UserID       int
	Enabled      bool
	PasswordHash string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Base64   string `json:"base64"`
}

// Bouncer is a middleware that checks if the user is authenticated.
func Bouncer(secretKey string) fiber.Handler {
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

func Base64Decode(b string) string {
	// Decode the base64 string
	base64Decoded, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		util.PrintError("Error decoding base64 string: " + err.Error())
	}

	// Print the decoded string
	util.PrintSuccess("Base64 decoded: " + string(base64Decoded))
	return string(base64Decoded)
}

// These functions are inspired from: https://github.com/go-webauthn/webauthn?tab=readme-ov-file#logging-into-an-account
func BeginLogin(c *fiber.Ctx) error { // TODO: Figure out if it's best to use the context, or the app instance

	var loginReq LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		util.PrintError("Error parsing login request: " + err.Error())
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request")
	}

	util.PrintDebug("Got a POST request to /login")
	username := loginReq.Username
	if username == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request")
	}

	util.PrintSuccess("Username: " + username)

	// If the base64 field is non-empty, we want to parse it:
	if loginReq.Base64 != "" {
		// Decode the base64 string
		base64Decoded, err := base64.StdEncoding.DecodeString(loginReq.Base64)
		if err != nil {
			util.PrintError("Error decoding base64 string: " + err.Error())
			return c.Status(fiber.StatusBadRequest).SendString("Invalid request")
		}

		// Print the decoded string
		util.PrintSuccess("Base64 decoded: " + string(base64Decoded))
	}

	user := GetUserByDisplayName(username)
	if user.UserID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
	}

	// Now we need to update last login time
	// UpdateLastLoginTime(username)

	// We also need to consider what to look for in the database when the user logs in.
	// It might be hard to remember their username and their three digits

	// Return a welcome message
	return c.Status(fiber.StatusOK).SendString("Welcome, " + username + "\nYour last login was " + user.LastLogin)

	// c.App().Post("/login", func(lctx *fiber.Ctx) error {
	// 	util.PrintDebug("Got a POST request to /login")
	// 	username := lctx.Params("username")
	// 	if username == "" {
	// 		return lctx.SendStatus(fiber.StatusBadRequest)
	// 	}

	// 	util.PrintSuccess("Username: " + username)
	// 	return c.Next()
	// })

	// General overview:
	// 1. Check if the user is already logged in by checking for a JWT token
	// (This is already done by the Bouncer)
	//   a) If the token is valid, just return a message to the hydrator(?) that the user is already logged in
	// 2. If the user is not logged in (i.e. the token is invalid), prompt the user to log in
	//   a) The username and password/fido2 credentials are received
	//   b) Fetch the user from the database to see if it exists
	// 	   i) If the user does not exist, jump to step 3
	//   c) If the user exists, check if the credentials are correct
	// 	   i) If the credentials are correct, generate a JWT token and send it back to the client.
	// 		  The frontend should store the token in the Authorization header for future requests
	// 	   ii) If the credentials are incorrect, jump to step 3
	// 3. If the credentials are invalid, return "Invalid credentials" to the client

	// 1: Check if the user is already logged in
}

// TODO: Implement the finish login process
func FinishLogin() {
	// Implement
}

// TODO: Implement the registration process
func BeginRegistration() {
	// Implement

}

// TODO: Implement the finish registration process
func FinishRegistration() {
	// Implement
}

// TODO: Implement the logoff process
func Logoff() {
	// General overview:
	// Find the user in the database and clear the session (token, increment the webauthn counter(?), etc.)
}

func EnableWebAuthn() {
	// Implement
}

func EnablePasswordAuth() {
	// Implement
}

func DisableWebAuthn() {
	// Implement
}

func DisablePasswordAuth() {
	// Implement
}
