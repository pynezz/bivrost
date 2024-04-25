package bivrost

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/fetcher"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/fswatcher"
	"github.com/pynezz/bivrost/internal/ipc/ipcserver"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/tui"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/flags"
	"github.com/pynezz/bivrost/modules"
)

// 1. The main function is the entry point of the application.
// 2. It initializes the tui.Header.Color and prints the header.
// 3. It checks if the number of arguments is less than 2 and prints a warning if it is.
// 4. It parses the flags.
// 5. Testing
// 5.1 If the test flag is set to "db", it tests the database connection.
// 5.2 It tests the UDS connection.
// 6. It loads the configuration file and connects to the database.
// 7. It initializes the Fiber server with the configuration values.
// 8. It sets the port to 3000 if the configuration file does not specify a port.

func Execute() {

	// Setting up signal handling to catch CTRL+C and other termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %s\n", sig)
		os.Exit(0)
	}()
	// Print the startup header
	tui.Header.Color = util.Cyan
	tui.Header.PrintHeader()

	// Check for command line arguments
	if len(os.Args) < 2 {
		util.PrintWarning("No arguments provided. Use -h for help.")

		flag.Usage()
		return
	}

	go func() {
		fswatcher.Watch("./access.log")
	}()

	// nginxDB, err := fetcher.ReadDB("logs")
	gormConf := gorm.Config{}
	nginxDB, err := database.NewDataStore[fetcher.NginxLog]("logs", gormConf)
	if err != nil {
		util.PrintError("Failed to read the database: " + err.Error())
		return
	}

	nginxDB.AutoMigrate()
	nginxDB.TestWrite("10k")
	nginx_log_test_001 := `{"time_local":"22/Apr/2024:17:56:07 +0000","remote_addr":"43.163.232.152","remote_user":"","request":"GET /viwwwsogou?op=8&query=%E7%A8%8F%E5%BB%BA%09%E9%BE%90%E1%B7%A2 HTTP/1.1","status": "400","body_bytes_sent":"248","request_time":"0.000","http_referrer":"","http_user_agent":"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko","request_body":"gorm test"}`

	parsedLog, err := fetcher.ParseNginxLog(nginx_log_test_001)
	if err != nil {
		if err != fetcher.EnvironError {
			util.PrintError("Failed to parse the log: " + err.Error())
		}
		util.PrintWarning("Log is an environment variable.")
	}

	nginxDB.InsertLog(parsedLog)
	util.PrintSuccess("Log inserted successfully.")
	// defer nginxDB.Close()

	// resultsDB, err := sql.Open("sqlite3", fetcher.ResultsDB)
	// if err != nil {
	// 	util.PrintError("Failed to open the results database: " + err.Error())
	// 	return
	// }
	// defer resultsDB.Close()

	// Parse the command line arguments (flags)
	flags.ParseFlags()

	// Load the config
	cfg, err := config.LoadConfig(*flags.Params.ConfigPath)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Exiting...")
		return
	}

	err = modules.LoadModules(*cfg)
	if err != nil {
		util.PrintError("Failed to load modules: " + err.Error())
		fmt.Println(err)
		return
	}

	// Testing db connection
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
	// go testProtoConnection()
	go testUDS()

	// Connect to database
	db, err := middleware.NewDBService().Connect(cfg.Database.Path)
	if err != nil {
		util.PrintError("Main function: " + err.Error())
		return
	}

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

	util.PrintItalic("[main.go] Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	util.PrintItalic("[main.go] Exiting...")
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

// Standard port: 50051
// func testProtoConnection() {
// 	connector.InitProtobuf(50051)
// }

func testUDS() {
	util.PrintInfo("Testing UNIX domain socket connection...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ipcServer := ipcserver.NewIPCServer("bivrost", "bivrost")
	ok := ipcServer.InitServerSocket()
	if !ok {
		return
	}

	// TODO: Check if this is more applicable: https://www.man7.org/linux/man-pages/man7/unix.7.html
	go ipcServer.Listen()

	util.PrintItalic("Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	<-c
	ipcserver.Cleanup()
	fmt.Println("Done cleaning up. Exiting...")
	// uds, err := connector.NewIPC("test", "Test socket")
	// if err != nil {
	// 	errorMsg := "main.go: could not connect to UNIX domain socket.\n" + err.Error()
	// 	util.PrintError(errorMsg)
	// 	return
	// }

	// util.PrintColor(util.BgCyan, "Connected to UNIX domain socket.")
	// uds.Initialize()

	// util.PrintColor(util.BgCyan, "Listening on UNIX domain socket...")

	// uds.Listen()
}
