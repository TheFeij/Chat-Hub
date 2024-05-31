package postgres

import (
	"Chat-Server/config"
	"os"
	"testing"
)

// TestMain performs initializations before tests are run and does cleanups after tests are run
func TestMain(m *testing.M) {
	conf := config.GetConfig("config", "json", "../../../config")

	// initialize the singleton instance of PostgresRepository with the address of the test database
	_ = GetPostgresRepository(conf.TestDatabaseAddress())

	code := m.Run()

	cleanupTestDatabase()

	os.Exit(code)
}

// cleanupDatabase removes all records from all tables in the database
func cleanupDatabase() {
	postgresRepository.db.Exec("DELETE FROM messages")
	postgresRepository.db.Exec("DELETE FROM sessions")
	postgresRepository.db.Exec("DELETE FROM users")
}
