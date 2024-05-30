package repository

import (
	"Chat-Server/repository/db/sql"
	"Chat-Server/repository/io"
)

// Repository implements the required methods for the business layer to interact with the data layer
type Repository interface {
	// AddMessage adds a message to the data layer
	AddMessage(message io.Message) (io.Message, error)

	// GetAllMessages retrieves all messages from the database
	GetAllMessages() ([]io.Message, error)
}

// NewRepository returns a new repository with the given type
func NewRepository(databaseType DatabaseType) Repository {
	switch databaseType {
	case Postgres:
		return sql.NewPostgresRepository()
	}

	return nil
}
