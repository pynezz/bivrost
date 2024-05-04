package database

import (
	"gorm.io/gorm"
)

// type Store[T any] interface {
// 	AutoMigrate(T any) error
// 	InsertLog(args ...any) error
// 	InitDB(database string) *gorm.DB
// 	GetAllLogs() ([]T, error)
// 	GetLogByIP(ip string) (T, error)
// 	GetByID(id uint) (*T, error)
// 	UpdateLog(updated T) error
// 	DeleteLog(id uint) error
// }

// Generic store interface
type Store[T any] interface {
	Name() string
	AutoMigrate() error
	InsertLog(log T) error
	GetAllLogs() ([]T, error)
	GetLogByID(id uint) (*T, error)
	GetLogRangeFromID(from int) ([]T, error)
	UpdateLog(log T) error
	DeleteLog(id uint) error
}

// Generic repository implementation
type DataStore[StoreType any] struct {
	name string
	db   *gorm.DB
	Type StoreType // ? Is this beneficial?
}

// The stores map is a map of store names to their respective DataStore
// that provides a generic interface to the database
var stores map[string]*DataStore[any]

// InitStores initializes the stores map
func init() {
	stores = make(map[string]*DataStore[any])
}

// AddStore adds a store to the stores map
func AddStore[T any](store *DataStore[any]) {
	stores[store.Name()] = store
}

// GetStore returns a store from the stores map
func GetStore(name string) *DataStore[any] {
	return stores[name]
}

func Import() {
	// Import all the stores
	// AddStore(NewDataStore[models.NginxLog](nginxLogDB, "nginx_logs"))
}
