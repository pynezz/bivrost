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
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // https://pkg.go.dev/github.com/mattn/go-sqlite3#section-readme

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
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
	driver *sql.DB
}

var instance *Database // The global database instance
var isConnected *bool  // The global database connection status

// https://gosamples.dev/sqlite-intro/

// NewDatabase creates a new database. It returns a pointer to the database.
func InitDatabaseDriver(db *sql.DB) *Database {
	util.PrintInfo("Initializing new database driver...")
	return &Database{
		driver: db,
	}
}

func NewDBService() *Database {
	return &Database{} // TODO: Implement
}

func GetDBInstance() (*Database, error) {
	// TODO: Add some error handling in case the instance is nil.
	if instance == nil || !*isConnected {
		return nil, error(util.Errorf(
			"Database instance is not connected, or is nil. Please connect to the database first via the Connect method."))
	}

	return instance, nil
}

// Connect to the database
func (db *Database) Connect(dbPath string) (*sql.DB, error) {
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
	// Open the database
	driver, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	*isConnected, err = db.IsConnected()
	if err != nil {
		return nil, err
	}

	db.driver = driver
	return driver, nil
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

		_, err = db.driver.Exec(string(query))
		if err != nil {
			return err
		}
	}
	return err
}

// Check for connectivity with the database
func (db *Database) IsConnected() (bool, error) {
	err := db.driver.Ping()
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
	result, err := db.driver.Exec(query, args...)
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

// TODO: Fetch executes a SELECT query and returns the result.
func (db *Database) Fetch(query string, args ...interface{}) ([]RowScanner, error) {
	rows, err := db.driver.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RowScanner
	for rows.Next() {
		var field RowScanner
		err := rows.Scan(&field)
		if err != nil {
			return nil, err
		}

		result = append(result, field)
		// Here we need to account for the db structure.
		// var field string
		// for field, i in rows.Scanner() { ... } // Something along these lines
		// result = append(result, &i)	// TODO: Do this properly
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func testWrite(database *Database) {
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

	// Generate a random 32-bit integer
	randomID := rand.Int31n(1000000)

	profileImageUrl := PlaceholderImage{
		Width:  200,
		Height: 200,
		Text:   displayName,
	}

	u := User{
		UserID:      int(randomID),
		DisplayName: displayName,
		// CreatedAt:   string(today),
		// UpdatedAt:   today,
		Role:            "user",
		FirstName:       "John",
		ProfileImageUrl: GetPlaceholderImage(profileImageUrl),
		// SessionId:       rand.Int(10, 32),
		AuthMethodID: 1,
	}

	err := database.Write("INSERT INTO users (UserID, DisplayName, CreatedAt, UpdatedAt, Role, FirstName, ProfileImageURL, SessionId, AuthMethodID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		u.UserID, u.DisplayName, u.CreatedAt, u.UpdatedAt, u.Role, u.FirstName, u.ProfileImageURL, u.SessionId, u.AuthMethodID)
	if err != nil {
		util.PrintError(err.Error())
	}

}

func testPrintRows(rows []RowScanner) {
	for _, row := range rows {
		fmt.Printf("%v\n", row)
	}
}
