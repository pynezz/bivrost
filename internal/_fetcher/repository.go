package fetcher

// Generic repository interface
type Repository[T any] interface {
	Create(args ...any) error
	All() ([]T, error)
	GetByID(id int64) (*T, error)
	Update(id int64, updated T) (*T, error)
	Delete(id int64) error
}
