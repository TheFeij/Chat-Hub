package repository

import (
	"Chat-Server/repository/db/sql"
)

// Repository implements the required methods for the business layer to interact with the data layer
type Repository interface {
	// TODO implement necessary methods for the business layer to interact with the database
}

// NewRepository returns a new repository with the given type
func NewRepository(databaseType DatabaseType) Repository {
	switch databaseType {
	case Postgres:
		return sql.NewPostgresRepository()
	}

	return nil
}
