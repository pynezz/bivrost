package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/connector"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/tui"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/flags"
)

// 1. The main function is the entry point of the application.
// 2. It initializes the tui.Header.Color and prints the header.
// 3. It checks if the number of arguments is less than 2 and prints a warning if it is.
// 4. It parses the flags.
// 5. If the test flag is set, it checks if the value is "db" and calls the testDbConnection function.
// 6. It loads the configuration file and prints it.
// 7. It creates a new server and listens on port 3000.

func main() {
	// Print the startup header
	tui.Header.Color = util.Cyan
	tui.Header.PrintHeader()

	// Check for command line arguments
	if len(os.Args) < 2 {
		util.PrintWarning("No arguments provided. Use -h for help.")

		flag.Usage()
		return
	}

	// Parse the command line arguments (flags)
	flags.ParseFlags()
	if *flags.Params.Test != "" {
		if *flags.Params.Test == "db" {
			testDbConnection()
			return
		} else {
			util.PrintError("Invalid test parameter. Exiting...")
			return
		}
	}

	// Testing the proto connection
	go testProtoConnection()

	// Load the config
	cfg, err := config.LoadConfig(*flags.Params.ConfigPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting...")
		return
	}

	// var argon2Obj = cryptoutils.Argon2{} // Change the type to cryptoutils.Argon2
	// cryptoutils.InitArgon2(&argon2Obj) // Pass the address of argon2Obj

	// Connect to database
	db, err := middleware.NewDBService().Connect(cfg.Database.Path)
	if err != nil {
		util.PrintError("Main function: " + err.Error())
		return
	}
	// database := &middleware.Database{
	// 	Driver: db,
	// }

	// middleware.TestWrite(middleware.GetDBInstance())

	// db, err := middleware.InitDatabaseDriver().Connect(cfg.Database.Path)
	// if err != nil {
	// 	util.PrintError("Main function: " + err.Error())
	// 	return
	// }

	// As stated in the documentation:
	// 	- It is rare to Close a DB, as the DB handle is meant to be long-lived and shared between many goroutines.
	// However this is a defer statement, so it will be called when the function returns, which is the end of the main function.
	// Meaning that the database will be closed when the application is closed.
	defer db.Driver.Close()

	port := 3000
	if cfg.Network.Port != 0 {
		port = cfg.Network.Port
	}
	// Create the web server
	app := api.NewServer(cfg)
	app.Listen(":" + strconv.Itoa(port))
}

const dbPath = "users.db" // Testing purposes. This should be in the config file

func testDbConnection() {
	if fsutil.FileExists(dbPath) {
		util.PrintWarning("Removing the existing database file...")
		if err := os.Remove(dbPath); err != nil {
			util.PrintError(err.Error())
			return
		}
		util.PrintSuccess("Database file removed.")
	}

	util.PrintInfo("Connecting to the database...")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	usersDb := middleware.InitDatabaseDriver(db)

	if err := usersDb.Migrate(); err != nil {
		util.PrintError(err.Error())
		return
	}

	util.PrintInfo("Testing db connection...")
	err = db.Ping()
	if err != nil {
		util.PrintError(err.Error())
		return
	}
	util.PrintColorBold(util.LightGreen, "🎉 Database connected!")
}

func testProtoConnection() {
	connector.Initialize()
}
