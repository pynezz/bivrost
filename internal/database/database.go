package database

/*
	I've resorted to using the gorm library to interact with the database.
	This is not the frontend auth store, but the backend module data store for data such as logs, and results.

	I'll now be referring to the database setups as "data store" to differentiate it from the database logic.
*/

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pynezz/bivrost/internal/database/models"
	"github.com/pynezz/bivrost/internal/util"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

const (
	LogsDB             = "logs"
	ResultsDB          = "results"
	nginx_log_test_001 = `{"time_local":"22/Apr/2024:17:56:07 +0000","remote_addr":"43.163.232.152","remote_user":"","request":"GET /viwwwsogou?op=8&query=%E7%A8%8F%E5%BB%BA%09%E9%BE%90%E1%B7%A2 HTTP/1.1","status": "400","body_bytes_sent":"248","request_time":"0.000","http_referrer":"","http_user_agent":"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko","request_body":""}`
	nginx_log_test_002 = `{"time_local":"22/Apr/2024:16:53:00 +0000","remote_addr":"91.90.40.176","remote_user":"","request":"HEAD /(/302.php HTTP/1.1","status": "404","body_bytes_sent":"0","request_time":"0.037","http_referrer":"","http_user_agent":"DirBuster-1.0-RC1 (http://www.owasp.org/index.php/Category:OWASP_DirBuster_Project)","request_body":""}`
	nginx_log_test_003 = `{"time_local":"22/Apr/2024:13:39:49 +0000","remote_addr":"91.90.40.176","remote_user":"","request":"POST /login HTTP/1.1","status": "302","body_bytes_sent":"0","request_time":"0.010","http_referrer":"http://164.92.132.240/","http_user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36","request_body":"username=admin&password=password_1234"}`
)

var (
	EnvironError = errors.New("log is an environment variable")

	dbNames = map[string]string{
		LogsDB:    "logs.db",
		ResultsDB: "results.db",
		// Remember to add new databases here if needed in the future
	}
)

// func Testrun() {
// 	util.PrintColorBold(util.DarkGreen, "Testing module data store connection...")
// 	dbConf := gorm.Config{} // No config for now
// 	db, err := InitDB("logs", dbConf)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	nginxLogDB, err := NewDataStore[models.NginxLog](db, "nginx_logs")

// 	// AddStore[models.NginxLog]((*DataStore[any])(nginxLogDB)) // Specify the type explicitly

// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	// Insert a log
// 	if err != nil {
// 		if err != EnvironError {
// 			fmt.Println(err)
// 		}
// 		util.PrintWarning("Log is an environment variable.")
// 	}

// 	if err := nginxLogDB.InsertLog(log); err != nil {
// 		fmt.Println(err)
// 	}
// }

func (s *DataStore[T]) Name() string {
	return s.name
}

func chkExt(database string) string {
	strParts := strings.Split(database, ".")
	if l := len(strParts); l > 1 && strParts[1] == "db" {
		return strings.Split(database, ".")[0]
	} else if l > 1 && strParts[1] != "db" {
		return ""
	}
	return database
}

func InitLogsDB(config ...gorm.Config) (*gorm.DB, error) {
	dbConf := gorm.Config{}
	if c := len(config); c != 0 {
		dbConf = config[0]
	}
	util.PrintInfo("Initializing logs database...")
	return InitDB(LogsDB, dbConf, &models.NginxLog{})
}

// Initialize the results database
// config is optional
func InitResultsDB(config ...gorm.Config) (*gorm.DB, error) {
	dbConf := gorm.Config{}
	if c := len(config); c != 0 {
		dbConf = config[0]
	}
	util.PrintInfo("Initializing results database...")
	resultsDB, err := InitDB(ResultsDB, dbConf,
		&models.SynTraffic{},
		&models.AttackType{},
		&models.IndicatorsLog{},
		&models.GeoLocationData{},
		&models.GeoData{})
	if err != nil {
		fmt.Println(err)
	}
	return resultsDB, err
}

// Initialize the database with the given name and configuration, and automigrate the given tables
func InitDB(database string, conf gorm.Config, tables ...interface{}) (*gorm.DB, error) {
	if _, ok := isValidDb(database); !ok && database != "" {
		return nil, fmt.Errorf("database name missing or invalid. Format: <name>.db or <name")
	} else {
		database = chkExt(database)
	}

	db, err := gorm.Open(sqlite.Open(database+".db"), &conf)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(tables...); err != nil {
		return nil, err
	}

	db = db.Session(&gorm.Session{CreateBatchSize: 100})

	return db, nil
}

// Initialize DB and automigrate given model
func NewDataStore[StoreType any](db *gorm.DB, name string) (*DataStore[StoreType], error) {
	store := &DataStore[StoreType]{db: db, name: name}
	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	return store, nil
}

// Automigrate given model to the database
func (s *DataStore[T]) AutoMigrate() error {
	var instance T
	return s.db.AutoMigrate(instance)
}

// InsertLog inserts a log into the database
func (s *DataStore[T]) InsertLog(log T) error {
	result := s.db.Create(&log)
	return result.Error
}

func (s *DataStore[T]) insertBatch(batch []models.NginxLog) int {
	result := s.db.Create(&batch)
	if result.Error != nil {
		util.PrintError("Failed to insert batch: " + result.Error.Error())
	}

	util.PrintSuccess("Inserted batch of size: " + fmt.Sprintf("%d", len(batch)) + " into table: " + s.name + " of type " + fmt.Sprintf("%T", batch[0]))

	resString := fmt.Sprintf("SQL: %s\n", result.Statement.SQL.String())
	util.PrintColorAndBg(util.LightGreen, util.BgGray, resString)

	return int(result.RowsAffected)
}

// SQL TEST:  select id, remote_addr, time_local from nginx_logs where remote_addr and time_local is not null LIMIT 10;
func (s *DataStore[T]) InsertBulk(logChan <-chan models.NginxLog) error {
	var count int64
	batchSize := 100
	buffer := make([]models.NginxLog, 0)
	done := make(chan struct{})
	counter := 0

	go func() {

		// TODO: Fix close channel so that we can insert the rest of the logs (no worries for now, it will still insert them when it reaches 100 logs in the buffer)
		defer close(done)
		for log := range logChan {
			counter++
			util.PrintColorAndBg(util.White, "\033[40m", "Inserting log: ")
			util.PrintColor(util.DarkCyan, string(log.ID))
			buffer = append(buffer, log)
			if counter%batchSize == 0 {
				util.PrintColorAndBg(util.White, util.BgRed, "buffer limit reached, inserting batch...")
				count += int64(s.insertBatch(buffer))
				buffer = make([]models.NginxLog, 0)
			}
		}
	}()

	go func() {
		<-done
		if len(buffer) > 0 {
			util.PrintColorAndBg(util.White, util.DarkYellow, "inserting remaining logs...")
			count += int64(s.insertBatch(buffer))
		}
		util.PrintInfo("InsertBulk() finished inserting logs, closing buffer.")
	}()

	return nil
}

// GetAllLogs returns all logs from the database
func (s *DataStore[T]) GetAllLogs() ([]T, error) {
	var logs []T
	result := s.db.Find(&logs)
	return logs, result.Error
}

// GetLatestLogs returns the logs with row ID greater than the given ID from the database
func (s *DataStore[T]) GetLogRangeFromID(from int) ([]T, error) {
	var logs []T
	result := s.db.Where("id > ?", from-1).Find(&logs)
	return logs, result.Error
}

// GetEntriesByIP returns all logs with the given
func (s *DataStore[T]) GetLogsByIP(ip string) ([]T, error) {
	var entries []T
	util.PrintInfo("Getting entries by IP:" + ip)
	result := s.db.Where("remote_addr = ?", ip).Find(&entries)
	util.PrintInfo("Entries found: " + fmt.Sprintf("%d", len(entries)))

	return entries, result.Error
}

func GetTableCount(db *gorm.DB, table string) (int64, error) {
	var count int64
	result := db.Table(table).Count(&count)
	return count, result.Error
}

// GetUniqueValues returns the unique values of a column in a table with optional column names
// to filter by. If no column names are provided, all columns are returned.
// The function returns the unique values and an error if any.
// Example usage: GetUniqueValues(db, "nginx_logs", "remote_addr")
// Example usage: GetUniqueValues(db, "nginx_logs", "remote_addr", "status")
func GetUniqueValues(db *gorm.DB, table string, column ...string) ([]string, error) {
	var values []string
	var result *gorm.DB
	if a := column[0]; a == "all" || a == "*" || len(column) < 1 { // Get all columns
		result = db.Table(table).Distinct("*").Find(&values)
	} else {
		result = db.Table(table).Select(column).Distinct(column).Find(&values)
	}

	return values, result.Error
}

// GetUniqueValuesByIP returns the unique values of a column in a table with optional column names
// select distinct remote_addr FROM nginx_logs WHERE remote_addr IS NOT NULL;
func GetUniqueValuesByIP(db *gorm.DB, column, table, ip string) ([]string, error) {
	var values []string
	result := db.Table(table).Select(column).Where("remote_addr = ?", ip).Group(column).Find(&values)
	return values, result.Error
}

func GetUniqueValuesLen(db *gorm.DB, column, table string) (int, error) {
	values, result := GetUniqueValues(db, column, table)
	return len(values), result
}

func GetValueFromTable(db *gorm.DB, value, table string) ([]string, error) {
	var columns []string
	result := db.Table(table).Select(value).Find(&columns)
	return columns, result.Error
}

// GetLogByID returns the log with the given ID
func (s *DataStore[T]) GetLogByID(id uint) (*T, error) {
	var log T
	result := s.db.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// UpdateLog updates the log with the given ID
func (s *DataStore[T]) UpdateLog(log T) error {
	result := s.db.Save(&log)
	return result.Error
}

// DeleteLog deletes the log with the given ID
func (s *DataStore[T]) DeleteLog(id uint) error {
	var instance T

	result := s.db.Delete(&instance, id)
	return result.Error
}

func isValidDb(database string) (string, bool) {
	database = chkExt(database)
	for _, db := range dbNames {
		if db == database+".db" {
			util.PrintSuccess("Database name is valid: " + database)
			return database, true
		}
	}
	util.PrintError("Database name is invalid: " + database)
	return "", false
}
