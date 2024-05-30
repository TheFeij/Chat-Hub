package repository

import "Chat-Server/repository/db/sql"

// NewRepository returns a new repository with the given type
func NewRepository(databaseType DatabaseType) Repository {
	switch databaseType {
	case Postgres:
		return sql.NewPostgresRepository()
	}

	return nil
}
