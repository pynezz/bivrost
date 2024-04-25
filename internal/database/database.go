package database

/*
	I've resorted to using the gorm library to interact with the database.
	This is not the frontend auth store, but the backend module data store for data such as logs, and results.

	I'll now be referring to the database setups as "data store" to differentiate it from the database logic.
*/

import (
	"fmt"

	"github.com/pynezz/bivrost/internal/fetcher"
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

func Testrun() {
	util.PrintColorBold(util.DarkGreen, "Testing module data store connection...")
	dbConf := gorm.Config{} // No config for now
	db, err := NewDataStore[fetcher.NginxLog](fetcher.LogsDB, dbConf)
	if err != nil {
		fmt.Println(err)
	}

	// Insert a log
	log, err := fetcher.ParseNginxLog(nginx_log_test_001)
	if err != nil {
		fmt.Println(err)
	}

	if err := db.InsertLog(log); err != nil {
		fmt.Println(err)
	}

}

// Generic repository implementation
type DataStore[T any] struct {
	db *gorm.DB
}

// Initialize DB and automigrate given model
func NewDataStore[T any](database string, conf gorm.Config) (*DataStore[T], error) {
	if database == "" {
		return nil, fmt.Errorf("database name is required")
	}

	db, err := gorm.Open(sqlite.Open(database+".db"), &conf)
	if err != nil {
		return nil, err
	}

	store := &DataStore[T]{db: db}
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

// GetAllLogs returns all logs from the database
func (s *DataStore[T]) GetAllLogs() ([]T, error) {
	var logs []T
	result := s.db.Find(&logs)
	return logs, result.Error
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
