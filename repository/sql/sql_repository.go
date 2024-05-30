package sql

import (
	"gorm.io/gorm"
	"sync"
)

// PostgresRepository implements Repository
type PostgresRepository struct {
	db *gorm.DB
}

// ensure PostgresRepository implements Repository interface
var _ = (*PostgresRepository)(nil)

// postgres is the singleton instance of SQLDatabase
var postgres PostgresRepository

// once is used to ensure the singleton instance is initialize once
var once sync.Once

// NewPostgresRepository returns a new PostgresRepository
func NewPostgresRepository() *PostgresRepository {
	once.Do(func() {
		// TODO initialize the singleton instance of the SQLRepository
	})

	// TODO return the singleton instance of the SQLRepository
	return nil
}

// TODO implement methods of the Repository interface
