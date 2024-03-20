package middleware

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pynezz/bivrost/internal/util"
)

// path: internal/middleware/users/users.go
/* USERS PACKAGE INFORMATION
Users package contains the user model and user related functions
The user model is used to represent a user in the system
The user related functions are used to interact with the user model in the database
The user model is not meant to be used for authentication, or marshalling, but rather for user management
*/

//  Users table
// CREATE TABLE users (
//     UserID INTEGER PRIMARY KEY,
//     DisplayName TEXT UNIQUE NOT NULL,
//     CreatedAt TEXT DEFAULT (datetime('now')),
//     UpdatedAt TEXT DEFAULT (datetime('now')),
//     LastLogin TEXT,
//     Role TEXT CHECK(Role IN ('admin', 'user')) DEFAULT 'user',
//     FirstName TEXT,
//     ProfileImageURL TEXT,
//     SessionId TEXT,
//     AuthMethodID INTEGER
// );

/* SQLite initialization script for the webauthn_auth and password_auth tables */
/* PATH: db/migrations/002_create_auth_tables.sql */
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

// I want to define a set of errors to return when different kinds of validation errors occur. These could be interfaces, or structs, or just strings. I'm not sure yet.
// UserValidationError is a type of error that occurs when a user validation fails
type UserValidationError struct {
	Description string
}

// Error makes UserValidationError implement the error interface.
func (e *UserValidationError) Error() string {
	return e.Description
}

// NewUserValidationError creates a new UserValidationError with the provided message.
func NewUserValidationError(message string) *UserValidationError {
	return &UserValidationError{Description: message}
}

// This user will be sent back to the server, and then to the client as a JSON object
type User struct {
	UserID          uint64 `json:"id"`
	DisplayName     string `json:"displayname"`
	CreatedAt       string `json:"createdat"`
	UpdatedAt       string `json:"updatedat"`
	LastLogin       string `json:"lastlogin"`
	Role            string `json:"role"`
	FirstName       string `json:"firstname"`
	ProfileImageUrl string `json:"profileimageurl"`
	SessionId       string `json:"sessionid"`
	AuthMethodID    int    `json:"authmethodid"`
}

// type User struct {
// 	UserID          uint64
// 	DisplayName     string
// 	CreatedAt       string
// 	UpdatedAt       string
// 	LastLogin       string
// 	Role            string
// 	FirstName       string
// 	ProfileImageUrl string
// 	SessionId       string
// 	AuthMethodID    int
// }

type Users struct {
	Users    []User `json:"users"`
	Marshall func() string
}

// For incoming requests to create a user
type CreateUserRequest struct {
	DisplayName     string `json:"displayname"`
	Password        string `json:"password"`
	Role            string `json:"role"`
	FirstName       string `json:"firstname"`
	ProfileImageUrl string `json:"profileimageurl"`
}

// What to answer when a user is created
type CreateUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    User   `json:"data"`
}

// NewUser creates a new user
// displayname: The display name of the user
// password: The password of the user
// role: The role of the user
// firstname: The first name of the user
// profileimageurl: The profile image URL of the user (NOT required)
// sessionid: The session ID of the user
// authmethodid: The authentication method ID of the user (0 = password, 1 = webauthn)
func NewUser(
	displayname string,
	password string,
	role string,
	firstname string,
	profileimageurl string,
) CreateUserRequest {
	if firstname == "" {
		firstname = displayname
	}

	return CreateUserRequest{ // Returns a copy of the User struct with the given values
		DisplayName:     displayname,
		Password:        password,
		Role:            role,
		FirstName:       firstname,
		ProfileImageUrl: profileimageurl,
	}
}

// ValidateUser validates the user struct and returns a boolean and a slice of errors if the validation fails
// It's meant to be used by the CreateUser function, not by the user, or client itself
// To be certain nothing is changed in the user struct, I'm passing a copy of the user struct
// TODO: Test
func ValidateNewUser(user CreateUserRequest) (bool, []error) {
	var err []error // In case we got multiple validation errors

	if user.DisplayName == "" {
		err = append(err, NewUserValidationError("Display name is required"))
		// I want to keep going with the validation, even if one of the fields is invalid
		// So I'm not returning here, but rather appending the error to the err slice
	}
	if user.Password == "" {
		err = append(err, NewUserValidationError("Password is required"))
	}
	if len(user.Password) < 12 {
		err = append(err, NewUserValidationError("Password must be at least 12 characters long"))
	}
	if user.Role != "admin" && user.Role != "user" {
		err = append(err, NewUserValidationError("Invalid role. Should be 'admin' or 'user'"))
	}
	if user.ProfileImageUrl == "" {
		p := PlaceholderImage{
			Width:  200,
			Height: 200,
			Text:   user.DisplayName,
		}
		user.ProfileImageUrl = GetPlaceholderImage(p)
	}

	return false, err
}

func (p *PasswordAuth) ValidatePasswordAuth() (bool, []error) {
	var err []error

	if p.UserID == 0 {
		err = append(err, NewUserValidationError("User ID is required"))
	}
	if p.Enabled != 0 && p.Enabled != 1 {
		err = append(err, NewUserValidationError("Enabled must be 0 or 1"))
	}
	if p.PasswordHash == "" {
		err = append(err, NewUserValidationError("Password hash is required"))
	}

	return false, err
}

// GET USERS
// - Get users by ID
// - Get users by display name
// - Get users by role
// Then pass it back to the middleware which serializes it to JSON and sends it back to the client

// UserQuery is an interface for querying users
// It is used to define a set of methods for querying users that can be implemented by different types of user queries
// This is useful for testing and for creating different types of user queries, such as a database user query and a server user query
type UserQuery interface {
	GetUserByID(id string) // Return generic type
	GetUserByDisplayName(displayname string) User
}

// Important to note: https://go.dev/doc/database/sql-injection
// TODO: Probably best to use something else. Important to keep in mind that this is a potential security risk
// If we get this ID from the client, we need to ensure its integrity by validating it in a JWT token, or something
// We'll see if it's necessary to implement this
func GetUserByID(id string) User {
	// Lookup user id in the database
	// Return the result as a User struct
	var user User

	// Query the database for the user
	instance := GetDBInstance()
	if instance.Driver == nil {
		return user
	}

	err := instance.Driver.QueryRow(instance.SelectColEq(UCId), id).
		Scan(
			&user.UserID, &user.DisplayName, &user.CreatedAt,
			&user.UpdatedAt, &user.LastLogin, &user.Role,
			&user.FirstName, &user.ProfileImageUrl,
			&user.SessionId, &user.AuthMethodID,
		)

	if err != nil {
		util.PrintError("GetUserByID: " + err.Error())
	}
	return user
}

// GetUserByDisplayName returns a user with the given display name
// Will return a user with id 0 if the user is not found
func GetUserByDisplayName(displayname string) User { // Displayname is used to login

	// Lookup user display name in the database
	// Return the result as a User struct
	var user User
	user.UserID = 0

	// Query the database for the user
	instance := GetDBInstance()
	if instance.Driver == nil {
		return user
	}
	// err := instance.Driver.QueryRow(
	// 	`SELECT UserID, DisplayName, CreatedAt,
	// 	UpdatedAt, LastLogin, Role,
	// 	FirstName, ProfileImageURL,
	// 	SessionId, AuthMethodID
	// 	FROM users WHERE DisplayName = ?`, displayname).Scan( // Would be nice to have a function for this
	// 	&user.UserID, &user.DisplayName, &user.CreatedAt,
	// 	&user.UpdatedAt, &user.LastLogin, &user.Role,
	// 	&user.FirstName, &user.ProfileImageUrl,
	// 	&user.SessionId, &user.AuthMethodID)

	// Will find any user that has the display name provided, and with any three digits
	// SELECT * FROM users WHERE DisplayName LIKE 'displayname%'

	util.PrintDebug("SelectCol: " + instance.SelectCol(UCDisplayName, UClike) + displayname + "#" + "%")

	err := instance.Driver.QueryRow(instance.SelectCol(UCDisplayName, UClike), displayname+"#"+"%").
		Scan(
			&user.UserID, &user.DisplayName, &user.CreatedAt,
			&user.UpdatedAt, &user.LastLogin, &user.Role,
			&user.FirstName, &user.ProfileImageUrl,
			&user.SessionId, &user.AuthMethodID,
		)

	if err != nil {
		util.PrintError("GetUserByDisplayName: " + err.Error())
		return user
	}

	// If the length is greater than 0 we have a user
	// If the length is greater than 1 we have multiple users with the same display name
	// Iterate over the results and return the first one
	// - Or you know what, the user should be able to remember their display name and their three digits
	// if res := results.Next(); len(res) > 1 {
	// 	err := results.Scan(
	// 		&user.UserID, &user.DisplayName, &user.CreatedAt,
	// 		&user.UpdatedAt, &user.LastLogin, &user.Role,
	// 		&user.FirstName, &user.ProfileImageUrl,
	// 		&user.SessionId, &user.AuthMethodID,
	// 	)

	// 	if err != nil {
	// 		util.PrintError("GetUserByDisplayName: " + err.Error())
	// 	}
	// }
	// err := results.Scan(
	// 	&user.UserID, &user.DisplayName, &user.CreatedAt,
	// 	&user.UpdatedAt, &user.LastLogin, &user.Role,
	// 	&user.FirstName, &user.ProfileImageUrl,
	// 	&user.SessionId, &user.AuthMethodID,
	// )
	// if err != nil {
	// 	util.PrintError("GetUserByDisplayName: " + err.Error())
	// }
	return user
}

func GetPasswordHash(userId uint64) (PasswordAuth, error) { // Should maybe return an Argon2 struct
	pwAuth := PasswordAuth{
		UserID:       userId,
		Enabled:      1,
		PasswordHash: "",
	}

	// SELECT PasswordHash FROM password_auth WHERE UserID = ?
	util.PrintDebug("Getting password hash for user with ID: " + strconv.FormatUint(userId, 10))
	instance := GetDBInstance()
	if instance.Driver == nil {
		return pwAuth, fmt.Errorf("Database driver is nil")
	}
	rows, err := instance.Fetch(instance.GetPasswordHashQuery(), userId)
	if err != nil {
		util.PrintError("GetPasswordHash: " + err.Error())
		return pwAuth, err
	}

	if rows.Next() {
		err = rows.Scan(&pwAuth.PasswordHash)
		if err != nil {
			util.PrintError("GetPasswordHash: " + err.Error())
			return pwAuth, err
		}
	}
	return pwAuth, nil
}

func GetUserAuth(user User, method AuthMethod) {
	if user.AuthMethodID == 1 { // Password
		// Get user by display name
		// Get password hash
		// Compare password hash
		// Return user
	}

	if user.AuthMethodID == 2 { // WebAuthn
		// Get user by display name
		// Get webauthn data
		// Compare webauthn data
		// Return user
	}

	if user.AuthMethodID == 3 { // Both
		// Get user by display name
		// Get password hash
		// Compare password hash
		// Get webauthn data
		// Compare webauthn data
		// Return user
	}

}

// Check https://placeholders.dev/ for more info
type PlaceholderImage struct {
	Width      int ``
	Height     int
	Text       string
	FontFamily string
	FontWeight string
	FontSize   string
	Dy         string
	BgColor    string
	TextColor  string
	TextWrap   bool
}

// This is a very ugly function.
// TODO: Check if there's a better way to parse the struct
// TODO: Check the url package
// GetPlaceholderImage returns a URL to a placeholder image
func GetPlaceholderImage(params PlaceholderImage) string {
	baseURL := "https://images.placeholders.dev/"

	// Use url.Values to properly encode query parameters
	v := url.Values{}

	// Reflect on the struct to dynamically read field names and values
	val := reflect.ValueOf(params)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		// Convert each field to a corresponding query parameter
		switch value.Kind() {
		case reflect.Int:
			if value.Int() != -1 { // We'll use -1 as a "null" value
				// The key needs to be lowercase hence the strings.ToLower(field.Name)
				// We can't set it in lowercase by default because it's exported, and public fields need to be capitalized in Go
				v.Add(strings.ToLower(field.Name), strconv.FormatInt(value.Int(), 10))
			}
		case reflect.String:
			if value.String() != "" {
				v.Add(strings.ToLower(field.Name), value.String())
			}
		case reflect.Bool:
			v.Add(strings.ToLower(field.Name), strconv.FormatBool(value.Bool()))
		}
	}

	// Append encoded query parameters to the base URL
	urlWithParams := baseURL + "?" + v.Encode()

	fmt.Println("Placeholder image URL: " + urlWithParams)

	return urlWithParams
}

// UpdateLastLoginTime updates the last login time of a user
// It returns the amount of affected rows, or an error
func UpdateLastLoginTime(userId uint64) (sql.Result, error) {
	util.PrintDebug("Updating last login time for user with ID: " + strconv.FormatUint(userId, 10))

	// UPDATE users SET LastLogin = datetime('now') WHERE DisplayName = ?
	// var instance Database
	instance := GetDBInstance()
	if instance.Driver == nil {
		return nil, fmt.Errorf("Database driver is nil")
	}

	util.PrintDebug("writing to database")
	result, err := instance.Write(
		fmt.Sprintf("UPDATE users SET LastLogin = '%s' WHERE UserID = ?", time.Now().Format("2006-01-02 15:04:05")), userId)

	if err != nil {
		util.PrintError("UpdateLastLoginTime: " + err.Error())
		return nil, err
	}

	return result, nil
}

func LoginSuccessHTML(u User, jwt string) string {
	return fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Login success</title>
		</head>
		<body>
		<h1>Login success</h1>
		<i>%s</i>
		<img src="%s" alt="Profile image" />
		<h2>Username: %s</h2>
		<h3>User ID: %d</h3>
		<h3>Full username: %s</h3>
		<h3>Role: %s</h3>
		<h3>Session ID: %s</h3>
		<h3>Auth method ID: %d</h3>
		<h4 font-family: monospace;>Last login: %s</h4>
		<h4 font-family: monospace;>Created at: %s</h4>
		<h4 font-family: monospace;>Updated at: %s</h4>
		<h4 font-family: monospace;>JWT Token: %s</h4>
		</body>`,
		time.Now().Format("2006-01-02 15:04:05"), u.ProfileImageUrl,
		u.DisplayName, u.UserID, u.FirstName, u.Role, u.SessionId,
		u.AuthMethodID, u.LastLogin, u.CreatedAt, u.UpdatedAt, jwt)
}
