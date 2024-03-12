package users

import "github.com/pynezz/bivrost/internal/middleware"

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

type User struct {
	ID              int    `json:"id"`
	Displayname     string `json:"displayname"`
	Password        string `json:"password"`
	Role            string `json:"role"`
	FirstName       string `json:"firstname"`
	ProfileImageUrl string `json:"profileimageurl"`
	SessionId       string `json:"sessionid"`
	AuthMethodID    int    `json:"authmethodid"`
}

type Users struct {
	Users []User `json:"users"`
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
	sessionid string,
	authmethodid int) User {
	return User{ // Returns a copy of the User struct with the given values
		Displayname:     displayname,
		Password:        password,
		Role:            role,
		FirstName:       firstname,
		ProfileImageUrl: profileimageurl,
		SessionId:       sessionid,
		AuthMethodID:    authmethodid,
	}
}

func ValidateUser(user *User) bool {
	switch {
	case user.Displayname == "":
		return false
	case user.Password == "":
		return false
	case user.Role == "":
		return false
	case user.FirstName == "":
		return false
	case user.ProfileImageUrl == "":
		return false
	case user.SessionId == "":
		return false
	case user.AuthMethodID < 0:
		return false
	default:
		return true
	}
}

// GET USERS
// - Get users by ID
// - Get users by display name
// - Get users by role
// Then pass it back to the middleware which serializes it to JSON and sends it back to the client

type UserQuery interface {
	GetUserByID(id string) // Return generic type
	GetUserByDisplayName(displayname string) User
}

func GetUserByID(id string) User {
	// Lookup user id in the database
	// Return the result as a User struct
	user := User{}

	// Query the database for the user
	instance := middleware.GetDBInstance()
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
