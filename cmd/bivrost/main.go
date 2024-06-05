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
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/pynezz/bivrost/internal/api"
	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/database"
	"github.com/pynezz/bivrost/internal/fswatcher"
	"github.com/pynezz/bivrost/internal/ipc/ipcserver"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/tui"
	"github.com/pynezz/bivrost/internal/util/flags"
	"github.com/pynezz/bivrost/modules"

	util "github.com/pynezz/pynezzentials"
	"github.com/pynezz/pynezzentials/ansi"
	"github.com/pynezz/pynezzentials/fsutil"

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

var dbCreateBatchSize = 100

func Execute(isPackage bool, buildVersion string) {
	// Parse the command line arguments (flags)
	if !isPackage {
		flag.Parse()
	}

	args := flags.ParseFlags()
	ansi.PrintInfo(" > Config path: " + *args.ConfigPath)
	ansi.PrintInfo(" > Log path: " + *args.LogPath)

	// Setting up signal handling to catch CTRL+C and other termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %s\n", sig)
		os.Exit(0)
	}()

	termiui := tui.NewTui(buildVersion)
	termiui.Header.Color = ansi.Cyan
	termiui.Header.PrintHeader()

	// Check for command line arguments
	if len(os.Args) < 2 {
		ansi.PrintWarning("No arguments provided. Use -h for help.")

		flag.Usage()
		return
	}

	gormConf := gorm.Config{
		PrepareStmt:     true,
		CreateBatchSize: dbCreateBatchSize,

		Logger: logger.Default.LogMode(logger.Silent),
	}

	lineChan := make(chan string, dbCreateBatchSize) // Buffer of 1000 lines
	dataChan := make(chan string, dbCreateBatchSize)

	ansi.PrintBold("Testing module data store connection...")

	s, err := stores.ImportAndInit(gormConf)
	if err != nil {
		fmt.Println(err)
	}

	// Random IP for testing: 143.110.222.166
	randIP := "143.110.222.166"
	idOneLog, err := s.Get("nginx_logs").NginxLogStore.GetLogsByIP(randIP)
	if err != nil {
		fmt.Println(err)
	}
	if len(idOneLog) == 0 {
		fmt.Println("No logs found for IP: ", randIP)
	} else {
		fmt.Println("ID 1 log: ", idOneLog[0].ID)
		fmt.Printf("with %d fields containing the ip\n", len(idOneLog))
	}

	// nginxLogPath := "/var/log/nginx/access.log"
	// Fetch and parse the logs
	logPath := "nginx_50.log"
	if *args.LogPath != "" {
		logPath = *args.LogPath
	}

	fmt.Println("analyzing log " + logPath)

	go logalyzer(dataChan, lineChan, logPath, s.NginxLogStore)

	ansi.PrintDebug("Config path: " + *args.ConfigPath)
	// Load the config
	cfg, err := config.LoadConfig(*flags.Params.ConfigPath)
	if err != nil {
		ansi.PrintError("[bivrost|main.go] " + err.Error() + "\nExiting...")
		return
	}

	err = modules.LoadModules(*cfg)
	if err != nil {
		ansi.PrintError("[bivrost|main.go] Failed to load modules: " + err.Error())
		return
	}

	// Testing db connection
	if *flags.Params.Test != "" {
		if *flags.Params.Test == "db" {
			testDbConnection()
			return
		} else {
			ansi.PrintError("Invalid test parameter. Exiting...")
			return
		}
	}

	// Run the IPC goroutine
	go unixDomainSockets()

	// Connect to database
	db, err := middleware.NewDBService().Connect(cfg.Database.Path)
	if err != nil {
		ansi.PrintError("Main function: " + err.Error())
		return
	}

	port := 3300
	if cfg.Network.Port != 0 {
		port = cfg.Network.Port
	}

	// Create the web server
	app := api.NewServer(cfg)
	app.Listen(":" + strconv.Itoa(port))

	ansi.PrintItalic("[main.go] Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	<-sigChan
	ansi.PrintItalic("[main.go] Exiting...")
	if err := db.Driver.Close(); err != nil {
		ansi.PrintItalic("[+] db driver closed")
		time.Sleep(time.Second * 3)
	}
}

const dbPath = "users.db" // Testing purposes. This should be in the config file

func testDbConnection() {
	if fsutil.FileExists(dbPath) {
		ansi.PrintWarning("Removing the existing database file...")
		if err := os.Remove(dbPath); err != nil {
			ansi.PrintError(err.Error())
			return
		}
		ansi.PrintSuccess("Database file removed.")
	}

	ansi.PrintInfo("Connecting to the database...")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	usersDb := middleware.InitDatabaseDriver(db)

	if err := usersDb.Migrate(); err != nil {
		ansi.PrintError(err.Error())
		return
	}

	ansi.PrintInfo("Testing db connection...")
	err = db.Ping()
	if err != nil {
		ansi.PrintError(err.Error())
		return
	}
	ansi.PrintColorBold(ansi.LightGreen, "🎉 Database connected!")
}

func unixDomainSockets() {
	ansi.PrintInfo("Testing UNIX domain socket connection...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ipcServer := ipcserver.NewIPCServer("bivrost", "bivrost")
	ok := ipcServer.InitServerSocket()
	if !ok {
		return
	}

	// Listen for connections
	go ipcServer.Listen()

	ansi.PrintItalic("Waiting for SIGINT or SIGTERM... Press Ctrl+C to exit.")
	<-c

	ipcServer.CloseConn()
	ipcserver.Cleanup()
	fmt.Println("Done cleaning up. Exiting...")
}

func logalyzer(data chan string, lineChan chan string, log string, nginxLogStore *database.DataStore[models.NginxLog]) {
	ansi.PrintInfo("Starting the file watcher...")
	var wg sync.WaitGroup

	go fswatcher.Watch(log, data)

	logChan := make(chan models.NginxLog)

	// Parses from data and inserts the parsed logs into the logChan
	go database.ParseBufferedNginxLog(data, logChan)

	// Inserts n logs (200 for now) logs from logChan at a time
	go nginxLogWorker(nginxLogStore, logChan, &wg)

	for line := range data {
		// ansi.PrintInfo("Received line: " + line)
		lineChan <- line
	}

	defer func() {
		close(lineChan)
		close(logChan)
		ansi.PrintSuccess("logalyzer cleaned channels")
	}()
}

// nginxLogWorker is a worker function that processes the parsed logs and inserts them into the database.
func nginxLogWorker(nginxLogStore *database.DataStore[models.NginxLog], logChan <-chan models.NginxLog, wg *sync.WaitGroup) {
	timestamp := util.UnixNanoTimestamp()
	var finalTime int64
	ansi.PrintBold("Processing parsed logs for storage...")
	if err := nginxLogStore.InsertBulk(logChan, 100); err != nil {
		ansi.PrintError("Failed to insert logs: " + err.Error())
	} else {
		ansi.PrintSuccess("Logs inserted successfully.")
	}

	finalTime = util.UnixNanoTimestamp()
	elapsed := finalTime - timestamp
	ansi.PrintSuccess("Created and inserted the logs in")
	ansi.PrintSuccess(fmt.Sprintf(" > %d µsec", elapsed/1000))
	ansi.PrintSuccess(fmt.Sprintf(" > %d msec", elapsed/1000000))
	ansi.PrintSuccess(fmt.Sprintf(" > %d sec", elapsed/1000000000))
	ansi.PrintSuccess(fmt.Sprintf(" > %d min", elapsed/1000000000/60))
}
