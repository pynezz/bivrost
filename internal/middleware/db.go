package middleware

/* Middleware
   This code serve as the database middleware for the application.
   It is responsible for handling all operations related to directly interfacing and managing the database connectivity.
   This includes creating a new database, connecting to an existing one, and checking for connectivity.

   The middleware should not be responsible for handling the business logic of the application.

   The database middleware is responsible for:
   - Creating a new database
   - Connecting to an existing database
   - Checking for connectivity with the database
   - Closing the database connection
   - Fetching data from the database

*/

// DB Structure

/* PATH: db/migrations/001_create_users_table.sql */

// /* SQLite initialization script for the users table */
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
//     AuthMethodID INTEGER   /* This is a foreign key to the auth_methods table,
//                               but we need to add it later due to the auth tables being created after this one */
// );

/* SQLite initialization script for the webauthn_auth and password_auth tables */

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

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3" // https://pkg.go.dev/github.com/mattn/go-sqlite3#section-readme

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/cryptoutils"
)

// 💡 When creating a new SQLite database or connection to an existing one, with the file name additional options can be given.
// This is also known as a DSN (Data Source Name) string.

// The filename of the database as stored on disk
// const DBFileName = "users.db"

type RowScanner interface {
	Scan(dest ...interface{}) error
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Tables struct {
	Users        string
	AuthMethods  string
	UserSessions string
	WebAuthnAuth string
	PasswordAuth string
}

var DefaultTables = Tables{
	Users:        "users",
	AuthMethods:  "auth_methods",
	UserSessions: "user_sessions",
	WebAuthnAuth: "webauthn_auth",
	PasswordAuth: "password_auth",
}

func (t *Tables) GetTables() []string {
	return []string{t.Users, t.AuthMethods, t.UserSessions, t.WebAuthnAuth, t.PasswordAuth}
}

// Database defines the structure of the database. We're using SQLite in our project.
type Database struct {
	Driver *sql.DB
}

var DBInstance *Database

// var instance *Database // The global database instance
var sqlInstance *sql.DB // The global database connection

// https://gosamples.dev/sqlite-intro/

// NewDatabase creates a new database. It returns a pointer to the database.
func InitDatabaseDriver(db *sql.DB) *Database {
	util.PrintInfo("Initializing new database driver...")
	DBInstance = &Database{
		Driver: db,
	}
	return DBInstance
}

func NewDBService() *Database {
	return &Database{
		Driver: nil,
	}
}

// GetDBInstance returns the global database instance
func GetDBInstance() *Database {
	if DBInstance == nil {
		util.PrintError("Database instance is not connected, or is nil. Please connect to the database first via the Connect method.")
		return nil
	}
	return DBInstance
}

// Connect to the database
func (db *Database) Connect(dbPath string) (*Database, error) {
	util.PrintInfo("Connecting to the database...")
	// Check if the database file exists
	if fsutil.FileExists(dbPath) {
		util.PrintSuccess("Database file found.")
	} else {
		util.PrintWarning("Database file not found. Creating a new one...")

		// Leverage migration scripts to create new database
		err := db.Migrate()
		if err != nil {
			return nil, err
		}
	}
	util.PrintDebug("Opening the database...")
	// Open the database
	driver, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	util.PrintSuccess("Database opened.")

	util.PrintDebug("Verifying connection...")
	err = driver.Ping()
	if err != nil {
		return nil, err
	}

	util.PrintSuccess("Connection verified.")

	// Set the global database instance
	DBInstance = &Database{
		Driver: driver,
	}

	db.Driver = driver

	util.PrintDebug("Testing write...")
	// TestWrite(DBInstance)

	// testPrintRows(DBInstance)
	return DBInstance, nil
}

// Use this function to create a new database from the migration scripts
func (db *Database) Migrate() error {
	migrationFiles, err := fsutil.GetFiles("./db/migrations/")
	util.PrintSuccess(fmt.Sprintf("Found %d migration files", len(migrationFiles)))

	if err != nil {
		return err
	}

	for _, file := range migrationFiles {
		util.PrintInfo("Migrating: " + file)
		query, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// Prompt the user to confirm the migration
		util.PrintItalic(string(query))
		util.PrintColorBold(util.Yellow, "Do you want to continue? (y/n) ")
		var response string
		fmt.Scanln(&response)
		if response != "y" {
			util.PrintWarning("Skipping migration...")
			continue
		}
		util.PrintInfo("Executing migration: " + file)

		_, err = db.Driver.Exec(string(query))
		if err != nil {
			return err
		}
	}
	return err
}

type Write map[string]func(interface{}) string

// Check for connectivity with the database
func (db *Database) IsConnected() (bool, error) {
	err := db.Driver.Ping()
	return err == nil, err
}

// Write executes an INSERT, UPDATE, or DELETE query.
// Example usage:
// err := db.Write("INSERT INTO users (name) VALUES ({struct})", "John Doe", {...struct})
//
//	if err != nil {
//		return err
//	}
func (db *Database) Write(query string, args ...interface{}) (sql.Result, error) {
	result, err := db.Driver.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (db *Database) Fetch(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.Driver.Query(query, args...)
	util.PrintInfo("Fetched rows: " + fmt.Sprint(rows))
	if err != nil {
		return nil, err
	}

	return rows, nil
}

var RandomFirstName []string = []string{
	"John",
	"Jane",
	"Michael",
	"Sarah",
	"David",
	"Emily",
	"Daniel",
	"Olivia",
	"James",
	"Emma",
	"Benjamin",
	"Isabella",
	"Lucas",
	"Sophia",
	"Alexander",
	"Mia",
	"William",
	"Charlotte",
}

var RandomLastName []string = []string{
	"Smith",
	"Johnson",
	"Williams",
	"Brown",
	"Jones",
	"Garcia",
	"Miller",
	"Davis",
	"Rodriguez",
	"Martinez",
	"Hernandez",
	"Lopez",
	"Gonzalez",
	"Wilson",
	"Anderson",
	"Thomas",
	"Taylor",
	"Moore",
	"Jackson",
}

func TestWrite(database *Database) {
	displayName := getRandomName()

	// Generate a cryptographically secure random 32-bit integer
	var randomID uint64
	randomID, err := cryptoutils.GenerateRandomInt(1000000, 10000000) // I know I shouldn't have hardcoded these...
	util.PrintDebug("Generated random ID: " + strconv.Itoa(int(randomID)))
	if err != nil {
		util.PrintError(err.Error())
	}

	today := time.Now().Format("01-02-2006 15:04:05")

	profileImageUrl := PlaceholderImage{
		Width:  200,
		Height: 200,
		Text:   displayName,
	}

	userId := randomID

	u := User{
		UserID:          userId,
		DisplayName:     displayName + fmt.Sprintf("%s%s", "#", strconv.FormatUint(userId, 10)[:3]),
		CreatedAt:       today,
		UpdatedAt:       today,
		LastLogin:       today,
		Role:            "user",
		FirstName:       "John",
		ProfileImageUrl: GetPlaceholderImage(profileImageUrl),
		SessionId:       "123",
		AuthMethodID:    0,
	}

	rowsAffected, err := database.Write(`INSERT INTO users (
		UserID, DisplayName, CreatedAt,
		UpdatedAt, LastLogin, Role,
		FirstName, ProfileImageURL,
		SessionId, AuthMethodID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.UserID, u.DisplayName, u.CreatedAt,
		u.UpdatedAt, u.LastLogin, u.Role,
		u.FirstName, u.ProfileImageUrl,
		u.SessionId, u.AuthMethodID)

	if err != nil {
		util.PrintError(err.Error())
	}

	util.PrintSuccess(fmt.Sprintf("Inserted %d rows", rowsAffected))

	testPrintRows(database)
}

func testPrintRows(db *Database) {
	util.PrintDebug("Fetching all users...")
	rows, err := db.Fetch("SELECT * FROM users")

	var users []User
	for rows.Next() {
		var user User

		/* INFO
		In Go, when working with the database/sql package to scan rows from a query result into a struct,
		you need to explicitly list each field of the struct as separate arguments to the Scan method.
		This is because Scan uses reflection to assign the column values to the variables you pass,
		and it needs to know the exact structure of where to put each column's data.*/

		// Scan copies the columns in the current row into the values pointed at by dest.
		// The number of values in dest must be the same as the number of columns in Rows.
		if err := rows.Scan(&user.UserID, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin, &user.Role, &user.FirstName, &user.ProfileImageUrl, &user.SessionId, &user.AuthMethodID); err != nil {
			util.PrintError("Error scanning rows: " + err.Error())
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		util.PrintError("Error fetching rows: " + err.Error())
	}
	defer rows.Close()

	if err != nil {
		fmt.Println("Error fetching users:", err)
		return
	}
	util.PrintInfo("Found users: " + fmt.Sprint(len(users)))

	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}
}

func getRandomName() string {
	randFNameIndex, err := cryptoutils.GenerateRandomInt(0, len(RandomFirstName))
	if err != nil {
		util.PrintError(err.Error())
	}
	randLNameIndex, err := cryptoutils.GenerateRandomInt(0, len(RandomLastName))
	if err != nil {
		util.PrintError(err.Error())
	}

	return RandomFirstName[randFNameIndex] + " " + RandomLastName[randLNameIndex]
}

// Just helpers to not have to remember the column names and to avoid typos
// UC = User Column. Column names in the user database. Not sure if this is the Go way of doing it
const (
	UCId           string = "UserID"
	UCDisplayName  string = "DisplayName"
	UCRole         string = "Role"
	UCFirstName    string = "FirstName"
	UCCreatedAt    string = "CreatedAt"
	UCUpdatedAt    string = "UpdatedAt"
	UCLastLogin    string = "LastLogin"
	UCProfileUrl   string = "ProfileImageURL"
	UCSessionId    string = "SessionId"
	UCAuthMethodId string = "AuthMethodID"

	// User column operators, maybe not necessary
	UCeq        string = "="
	UCneq       string = "!="
	UClt        string = "<"
	UClte       string = "<="
	UCgt        string = ">"
	UCgte       string = ">="
	UClike      string = "LIKE"
	UCin        string = "IN"
	UCnotin     string = "NOT IN"
	UCand       string = "AND"
	UCor        string = "OR"
	UCnot       string = "NOT"
	UCisnull    string = "IS NULL"
	UCisnotnull string = "IS NOT NULL"
)

// To be used in conjuction with a QueryRow such that the column is passed as a parameter,
// and the value is passed as another parameter
// Example: SELECT * FROM users WHERE <column> = ?
// The first ? is the column, and the second ? is the value
//
// Query:
//
// SELECT UserID, DisplayName, CreatedAt, UpdatedAt, LastLogin, Role,
// FirstName, ProfileImageURL,
// SessionId, AuthMethodID
//
// FROM users WHERE %s = ?
func (d *Database) SelectColEq(col string) string {
	return fmt.Sprintf(`SELECT UserID, DisplayName, CreatedAt, UpdatedAt, LastLogin, Role,
		FirstName, ProfileImageURL,
		SessionId, AuthMethodID
		FROM users WHERE %s = ?`, col)
}

// Select a user column by the column name and the operator
// Example: SELECT * FROM users WHERE <column> <operator> ?
// (e.g., SELECT * FROM users WHERE [UserID] [=] ?)
func (d *Database) SelectCol(col string, operator string) string {
	return fmt.Sprintf(`SELECT UserID, DisplayName, CreatedAt, UpdatedAt, LastLogin, Role,
		FirstName, ProfileImageURL,
		SessionId, AuthMethodID
		FROM users WHERE %s %s ?`, col, operator)
}

func (d *Database) UpdateCell(col string) string {
	return fmt.Sprintf(`UPDATE users SET %s = ? WHERE UserID = ?`, col)
}

// Select a table by the column name and the value
func (d *Database) SelectFromPasswordAuth(col string, value string) string {
	return fmt.Sprintf(`SELECT UserID, Enabled, PasswordHash
		FROM password_auth WHERE %s = ?`, col)
}

// Select a table by the column name and the value
func (d *Database) SelectFromWebAuthnAuth(col string, value string) string {
	return fmt.Sprintf(`SELECT CredentialID, UserID, PublicKey, UserHandle, SignatureCounter, CreatedAt
		FROM webauthn_auth WHERE %s = ?`, col)
}

func (d *Database) SetAndEnablePasswordAuth(userId int, passwordHash string) string {
	return fmt.Sprintf(`INSERT INTO password_auth (UserID, Enabled, PasswordHash)
		VALUES (?, 1, ?)`, userId, passwordHash)
}

func (d *Database) SetWebAuthnAuth(credentialId string, userId int, publicKey string, userHandle string, signatureCounter int) string {
	return fmt.Sprintf(`INSERT INTO webauthn_auth (CredentialID, UserID, PublicKey, UserHandle, SignatureCounter)
		VALUES (?, ?, ?, ?, ?)`, credentialId, userId, publicKey, userHandle, signatureCounter)
}

func (d *Database) UpdateWebAuthnAuth(credentialId string, userId int, publicKey string, userHandle string, signatureCounter int) string {
	return fmt.Sprintf(`UPDATE webauthn_auth SET PublicKey = ?, UserHandle = ?, SignatureCounter = ?
		WHERE CredentialID = ? AND UserID = ?`, publicKey, userHandle, signatureCounter, credentialId, userId)
}

func (d *Database) UpdatePasswordAuth(userId int, passwordHash string) string {
	return fmt.Sprintf(`UPDATE password_auth SET PasswordHash = ?
		WHERE UserID = ?`, passwordHash, userId)
}

func (d *Database) GetPasswordHashQuery() string {
	return fmt.Sprintln(`SELECT PasswordHash FROM password_auth WHERE UserID = ?`)
}
