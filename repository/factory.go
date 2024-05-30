package repository

import "Chat-Server/repository/db/postgres"

// NewRepository returns a new repository with the given type
func NewRepository(databaseType DatabaseType) Repository {
	switch databaseType {
	case Postgres:
		return postgres.GetPostgresRepository()
	}

	return nil
}
