package api

import (
	"Chat-Server/config"
	"Chat-Server/repository"
	"Chat-Server/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var testConfigs *config.Config

// NewTestServer returns a new test server
func NewTestServer(t *testing.T, repository repository.Repository, tokenMaker token.Maker) *server {
	server := NewServer(repository, tokenMaker, testConfigs)
	require.NotEmpty(t, server)

	return server
}

// TestMain performs actions before and after tests
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	testConfigs = config.GetConfig("config", "json", "../config")

	exitCode := m.Run()
	os.Exit(exitCode)
}
