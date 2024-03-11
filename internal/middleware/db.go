package middleware

// This code serve as the database middleware for the application.

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/util"
)

// 💡 When creating a new SQLite database or connection to an existing one, with the file name additional options can be given.
// This is also known as a DSN (Data Source Name) string.

// The filename of the database as stored on disk
// const DBFileName = "users.db"

// Database defines the structure of the database. We're using SQLite in our project.
type Database struct {
	driver *sql.DB
}

// https://gosamples.dev/sqlite-intro/

// NewDatabase creates a new database. It returns a pointer to the database.
func InitDatabaseDriver(db *sql.DB) *Database {
	util.PrintInfo("Initializing new database driver...")
	return &Database{
		driver: db,
	}
}

func NewDBService() *Database {
	return &Database{}
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
	db.driver = driver
	return driver, nil

}

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
