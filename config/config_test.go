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

	require.Equal(t, "0.0.0.0:8080", conf.serverAddress)
	require.Equal(t, false, conf.isProductionEnv)
	require.Equal(t, "postgresql://root:1234@localhost:5432/chat_server?sslmode=disable", conf.databaseAddress)
	require.Equal(t, "postgresql://root:1234@localhost:5432/chat_server_test?sslmode=disable", conf.testDatabaseAddress)
	require.Equal(t, "********************************", conf.tokenSymmetricKey)
	require.Equal(t, 15*time.Minute, conf.accessTokenDuration)
	require.Equal(t, 24*time.Hour, conf.refreshTokenDuration)
	require.Equal(t, "/api/chat", conf.accessTokenCookiePath)
	require.Equal(t, "/api/refresh", conf.refreshTokenCookiePath)
	require.Equal(t, "/chat", conf.usernameCookiePath)
}
