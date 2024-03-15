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

func GetDBInstance() *Database {
	// TODO: Add some error handling in case the instance is nil.
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
	TestWrite(DBInstance)

	testPrintRows(DBInstance)
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
func (db *Database) Write(query string, args ...interface{}) error {
	result, err := db.Driver.Exec(query, args...)
	if err != nil {
		return err
	}

	// You can check the result (e.g., number of rows affected) if needed.
	// For example:
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

// // TODO: Fetch executes a SELECT query and returns the result.
// func (db *Database) Fetch(query string, args ...interface{}) ([]RowScanner, error) {
// 	rows, err := db.Driver.Query(query, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var result []RowScanner
// 	for rows.Next() {
// 		var field RowScanner
// 		err := rows.Scan(&field)
// 		if err != nil {
// 			return nil, err
// 		}

// 		result = append(result, field)
// 		// Here we need to account for the db structure.
// 		// var field string
// 		// for field, i in rows.Scanner() { ... } // Something along these lines
// 		// result = append(result, &i)	// TODO: Do this properly
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }

func (db *Database) Fetch(query string, args ...interface{}) ([]*User, error) {
	rows, err := db.Driver.Query(query, args...)
	util.PrintInfo("Fetched rows: " + fmt.Sprint(rows))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
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
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func TestWrite(database *Database) {
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
	//
	// today := time.Now().Format("01-02-2006 15:04:05")
	displayName := "John Doe"

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
		DisplayName:     displayName + fmt.Sprintf("%s%s", ":", strconv.FormatUint(userId, 10)[:3]),
		CreatedAt:       today,
		UpdatedAt:       today,
		LastLogin:       today,
		Role:            "user",
		FirstName:       "John",
		ProfileImageUrl: GetPlaceholderImage(profileImageUrl),
		SessionId:       "123",
		AuthMethodID:    0,
	}

	err = database.Write(`INSERT INTO users (
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

	testPrintRows(database)
}

func testPrintRows(db *Database) {
	util.PrintDebug("Fetching all users...")
	users, err := db.Fetch("SELECT * FROM users")
	if err != nil {
		fmt.Println("Error fetching users:", err)
		return
	}
	util.PrintInfo("Found users: " + fmt.Sprint(len(users)))

	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}
}
