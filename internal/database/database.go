package database

import (
	"github.com/pynezz/bivrost/internal/fetcher"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

func testrun() {
	ng := &fetcher.NginxLog{}

	db := NewDataStore[fetcher.NginxLog](fetcher.LogsDB)
	db.AutoMigrate(*ng)

}

// Generic repository implementation
type DataStore[T any] struct {
	db *gorm.DB
}

// Initialize DB and automigrate given model
func NewDataStore[T any](database string) *DataStore[T] {
	db, err := gorm.Open(sqlite.Open(database+".db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	store := &DataStore[T]{db: db}
	store.AutoMigrate()
	return store
}

func (s *DataStore[T]) AutoMigrate(instance T) error {
	return s.db.AutoMigrate(instance)
}

func (s *DataStore[T]) InsertLog(log T) error {
	result := s.db.Create(&log)
	return result.Error
}

func (s *DataStore[T]) GetAllLogs() ([]T, error) {
	var logs []T
	result := s.db.Find(&logs)
	return logs, result.Error
}

func (s *DataStore[T]) GetLogByID(id uint) (*T, error) {
	var log T
	result := s.db.First(&log, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

func (s *DataStore[T]) UpdateLog(log T) error {
	result := s.db.Save(&log)
	return result.Error
}

// TODO: Check this closer.
func (s *DataStore[T]) DeleteLog(id uint) error {
	result := s.db.Delete(&T{}, id)
	return result.Error
}
