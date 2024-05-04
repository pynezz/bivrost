package bivrost

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/fsutil"
	"github.com/pynezz/bivrost/internal/fswatcher"
	"github.com/pynezz/bivrost/internal/ipc/ipcserver"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/tui"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/flags"
	"github.com/pynezz/bivrost/modules"

	"github.com/pynezz/bivrost/internal/database/models"
	"github.com/pynezz/bivrost/internal/database/stores"
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

	termiui := tui.NewTui()
	termiui.Header.Color = util.Cyan
	// go termiui.Display()
	// Create a channel to receive the log data
	// termiui.AddDataSource(dataChan, "Watcher Output", util.Yellow) // Adding the file watcher output as a data source to the TUI

	// Check for command line arguments
	if len(os.Args) < 2 {
		util.PrintWarning("No arguments provided. Use -h for help.")

		flag.Usage()
		return
	}

	// util.PrintDebug("Testing Sigma rules...")
	// sigma.Test()

	// nginxDB, err := fetcher.ReadDB("logs")
	gormConf := gorm.Config{
		PrepareStmt:     true,
		CreateBatchSize: 100,

		Logger: logger.Default.LogMode(logger.Info),
	}

	// // logs.db
	// logsDatabase, err := database.InitLogsDB(gormConf)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// // results.db
	// modulesData, err := database.InitResultsDB(gormConf)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	lineChan := make(chan string, 1000) // Buffer of 1000 lines
	// logChan := make(chan models.NginxLog, 1000) // Buffer of 1000 logs
	dataChan := make(chan string, 1000)

	util.PrintBold("Testing module data store connection...")

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// synTrafficRepo, _ := database.NewDataStore[models.SynTraffic](modulesData, "syntraffic")
	// attackTypeRepo, _ := database.NewDataStore[models.AttackType](modulesData, "attacktype")
	// indicatorsLogRepo, _ := database.NewDataStore[models.IndicatorsLog](modulesData, "indicatorslog")
	// geoLocationDataRepo, _ := database.NewDataStore[models.GeoLocationData](modulesData, "geolocationdata")
	// geoDataRepo, _ := database.NewDataStore[models.GeoData](modulesData, "geodata")
	// nginxLogStore, err := database.NewDataStore[models.NginxLog](logsDatabase, "nginx_logs")

	// util.PrintSuccess("Nginx log data store connection successful: " + modulesData.Name())
	// util.PrintSuccess("Module data store connection successful: " + logsDatabase.Name())

	s, err := stores.ImportAndInit(gormConf)
	if err != nil {
		fmt.Println(err)
	}

	idOneLog, err := s.Get("nginx_logs").NginxLogStore.GetLogsByIP("")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("ID 1 log: ", idOneLog)

	// nginxLogPath := "/var/log/nginx/access.log"
	// Fetch and parse the logs
	go logalyzer(dataChan, lineChan, "/home/xkali/standard.log", s.NginxLogStore)
	// nginxLogWorker(nginxLogStore, lineChan, logChan)

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

	<-sigChan

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

func logalyzer(data chan string, lineChan chan string, log string, nginxLogStore *database.DataStore[models.NginxLog]) {
	util.PrintInfo("Starting the file watcher...")
	var wg sync.WaitGroup

	go fswatcher.Watch(log, data)

	logChan := make(chan models.NginxLog)
	go database.ParseBufferedNginxLog(data, logChan)
	go nginxLogWorker(nginxLogStore, logChan, &wg)

	for line := range data {
		util.PrintInfo("Received line: " + line)
		lineChan <- line
	}

	defer func() {
		// wg.Wait()
		close(lineChan)
		close(logChan)
	}()

	// go func() {
	// 	for {
	// 		select {
	// 		case line := <-data:
	// 			util.PrintInfo("Received line: " + line)
	// 			lineChan <- line

	// 			modelsChan := make(chan models.NginxLog, len(data))
	// 			go database.ParseBufferedNginxLog(lineChan, modelsChan)
	// 			nginxLogWorker(nginxLogStore, lineChan, modelsChan)

	// 		}
	// 	}
	// }()
}

// nginxLogWorker is a worker function that processes the parsed logs and inserts them into the database.
func nginxLogWorker(nginxLogStore *database.DataStore[models.NginxLog], logChan <-chan models.NginxLog, wg *sync.WaitGroup) {
	timestamp := util.UnixNanoTimestamp()
	var finalTime int64
	util.PrintBold("Processing parsed logs for storage...")
	if err := nginxLogStore.InsertBulk(logChan); err != nil {
		util.PrintError("Failed to insert logs: " + err.Error())
	} else {
		util.PrintSuccess("Logs inserted successfully.")

	}

	util.PrintBold("Processing parsed logs for storage...")
	if err := nginxLogStore.InsertBulk(logChan); err != nil {
		util.PrintError("Failed to insert logs: " + err.Error())
	} else {
		util.PrintSuccess("Logs inserted successfully.")
	}

	// for log := range logChan {
	// 	util.PrintInfo("Processing parsed log for storage")
	// 	if err := nginxLogStore.InsertLog(log); err != nil {
	// 		util.PrintError("Failed to insert log: " + err.Error())
	// 	}
	// }
	// go func() {
	// 	for {
	// 		select {
	// 		case log := <-lineChan:
	// 			util.PrintInfo("Received log: " + log)
	// 			parsedLog, err := database.ParseNginxLog(log)
	// 			if err != nil {
	// 				util.PrintError("Failed to parse log: " + log)
	// 				continue
	// 			}

	// 			nginxLogStore.InsertLog(parsedLog)
	// 		}
	// 		// close(logChan)
	// 	}
	// }()
	// util.PrintInfo("Waiting for the inserts to complete...")
	// close(logChan)

	finalTime = util.UnixNanoTimestamp()
	elapsed := finalTime - timestamp
	util.PrintSuccess(fmt.Sprintf("Created 10k logs\n > %d µsec", elapsed/1000))
	util.PrintSuccess(fmt.Sprintf(" > %d msec", elapsed/1000000))
	util.PrintSuccess(fmt.Sprintf(" > %d sec", elapsed/1000000000))
	util.PrintSuccess(fmt.Sprintf(" > %d min", elapsed/1000000000/60))
}
