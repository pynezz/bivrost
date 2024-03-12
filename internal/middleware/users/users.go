package users

/**
CREATE TABLE users (
    UserID INTEGER PRIMARY KEY AUTOINCREMENT,
    DisplayName TEXT UNIQUE NOT NULL,
    CreatedAt TEXT DEFAULT (datetime('now')),
    UpdatedAt TEXT DEFAULT (datetime('now')),
    LastLogin TEXT,
    Role TEXT CHECK(Role IN ('admin', 'user')) DEFAULT 'user',
    FirstName TEXT,
    ProfileImageURL TEXT,
    SessionId TEXT,
    AuthMethodID INTEGER
);
**/

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
