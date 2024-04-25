package database

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
	AutoMigrate() error
	InsertLog(log T) error
	GetAllLogs() ([]T, error)
	GetLogByID(id uint) (*T, error)
	UpdateLog(log T) error
	DeleteLog(id uint) error
}
