package config

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestGetConfig tests GetConfig function
func TestGetConfig(t *testing.T) {
	conf := GetConfig("config_test", "json", ".")
	require.NotEmpty(t, conf)

	require.Equal(t, "SERVER_ADDRESS", conf.serverAddress)
	require.Equal(t, "DATABASE_ADDRESS", conf.databaseAddress)
	require.Equal(t, "TEST_DATABASE_ADDRESS", conf.testDatabaseAddress)
	require.Equal(t, "TOKEN_SYMMETRIC_KEY", conf.tokenSymmetricKey)
	require.Equal(t, 15*time.Minute, conf.accessTokenDuration)
	require.Equal(t, 24*time.Hour, conf.refreshTokenDuration)
}
