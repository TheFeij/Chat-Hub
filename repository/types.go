package repository

// DatabaseType represents all supported databases
// this type is used as an input for the NewRepository() factory function
type DatabaseType string

const Postgres DatabaseType = "postgres"
