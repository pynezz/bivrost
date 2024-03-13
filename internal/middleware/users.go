package middleware

import (
	"strconv"

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
//     UserID INTEGER PRIMARY KEY AUTOINCREMENT,
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

// I need a separate kind of user for this package,
// because the user in the auth package is used for authentication,
// and this user is used for user management
type User struct {
	UserID          int    `json:"id"` // As of now (13.03.24), this is set by the database. Should be set by bivrost
	DisplayName     string `json:"displayname"`
	LastLogin       string `json:"lastlogin"`
	Role            string `json:"role"`
	FirstName       string `json:"firstname"`
	ProfileImageUrl string `json:"profileimageurl"`
	SessionId       string `json:"sessionid"`
	AuthMethodID    int    `json:"authmethodid"`
}

type Users struct {
	Users []User `json:"users"`
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
// profileimageurl: The profile image URL of the user
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

	switch {
	case user.DisplayName == "":
		err = append(err, NewUserValidationError("Display name is required"))
		// I want to keep going with the validation, even if one of the fields is invalid
		// So I'm not returning here, but rather appending the error to the err slice

	case len(user.Password) < 12:
		err = append(err, NewUserValidationError("Password must be at least 12 characters long"))
	case user.Role != "admin" && user.Role != "user":
		err = append(err, NewUserValidationError("Invalid role. Should be 'admin' or 'user'"))
	case user.ProfileImageUrl == "":
		err = append(err, NewUserValidationError("Profile image URL is required"))
	default:
		return true, nil
	}
	return false, err
}

func ValidateUserPasswordAuth(userAuth PasswordAuth) (bool, []error) {
	var err []error

	switch {
	case userAuth.PasswordHash == "":
		err = append(err, NewUserValidationError("Password is required"))

	// case len(userAuth.PasswordHash.Password) < 12:
	// err = append(err, NewUserValidationError("Password must be at least 12 characters long"))

	default:
		return true, nil
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
func GetUserByID(id string) User {
	// Lookup user id in the database
	// Return the result as a User struct
	user := User{}

	// Query the database for the user
	instance, err := GetDBInstance()
	if err != nil {
		util.PrintErrorf("Something has went wrong in the application control flow. Check if the database is ever connected.\nExiting...\n")
	}
	instance.Fetch("SELECT * FROM users WHERE UserID = ?", id)
	return user
}

func GetUserByDisplayName(displayname string) User {
	// Lookup user display name in the database
	// Return the result as a User struct

	user := User{}
	return user
}

func DbToUserStruct() {

}

// https://placeholders.dev/
//
// Available API Options
// width 		- Width of generated image. Defaults to 300.
// height 		- Height of generated image. Defaults to 150.
// text 		- Text to display on generated image. Defaults to the image dimensions.
// fontFamily 	- Font to use for the text. Defaults to sans-serif.
// fontWeight 	- Font weight to use for the text. Defaults to bold.
// fontSize 	- Font size to use for the text. Defaults to 20% of the shortest image dimension, rounded down.
// dy 			- Adjustment applied to the dy attribute of the text element to appear vertically centered. Defaults to 35% of the font size.
// bgColor 		- Background color of the image. Defaults to #ddd
// textColor 	- Color of the text. For transparency, use an rgba or hsla value. Defaults to rgba(0,0,0,0.5)
// textWrap 	- Wrap text to fit within the image (to best ability). Will not alter font size, so especially long strings may still appear outside of the image. Defaults to false
// Example URL
// https://images.placeholders.dev/?width=1055&height=100&text=Made%20with%20placeholders.dev&bgColor=%23f7f6f6&textColor=%236d6e71

type PlaceholderImage struct {
	Width      int
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

// GetPlaceholderImage returns a URL to a placeholder image
func GetPlaceholderImage(params PlaceholderImage) string {
	// Make a request to the placeholders.dev API
	// Return the image as a byte array

	url := "https://images.placeholders.dev/"

	switch {
	case params.Width != 0:
		url += "?width=" + string(params.Width)
	case params.Height != 0:
		url += "&height=" + string(params.Height)
	case params.Text != "":
		url += "&text=" + params.Text
	case params.FontFamily != "":
		url += "&fontFamily=" + params.FontFamily
	case params.FontWeight != "":
		url += "&fontWeight=" + params.FontWeight
	case params.FontSize != "":
		url += "&fontSize=" + params.FontSize
	case params.Dy != "":
		url += "&dy=" + params.Dy
	case params.BgColor != "":
		url += "&bgColor=" + params.BgColor
	case params.TextColor != "":
		url += "&textColor=" + params.TextColor
	case params.TextWrap:
		url += "&textWrap=" + strconv.FormatBool(params.TextWrap)
	default:
		url += ""
	}

	return url
}
