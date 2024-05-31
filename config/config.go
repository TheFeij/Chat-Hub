package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// config holds configuration variables
type config struct {
	databaseAddress      string        // address of the database
	testDatabaseAddress  string        // address of the test database
	serverAddress        string        // address of the server
	tokenSymmetricKey    string        // symmetric key of to make and verify tokens
	accessTokenDuration  time.Duration // access token duration
	refreshTokenDuration time.Duration // refresh token duration
}

// TokenSymmetricKey returns token symmetric key
func (c config) TokenSymmetricKey() string {
	return c.tokenSymmetricKey
}

// AccessTokenDuration returns access token duration
func (c config) AccessTokenDuration() time.Duration {
	return c.accessTokenDuration
}

// RefreshTokenDuration returns refresh token duration
func (c config) RefreshTokenDuration() time.Duration {
	return c.refreshTokenDuration
}

// DatabaseAddress returns database address
func (c config) DatabaseAddress() string {
	return c.databaseAddress
}

// TestDatabaseAddress returns test database address
func (c config) TestDatabaseAddress() string {
	return c.testDatabaseAddress
}

// ServerAddress returns server address
func (c config) ServerAddress() string {
	return c.serverAddress
}

// GetConfig returns a config object loaded with the config variables of the
// file specified in the input
func GetConfig(configFileName, configFileType, configFilePath string) *config {
	viper.AddConfigPath(configFilePath)
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)

	// read environment variables
	viper.AutomaticEnv()

	// read configurations
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("unable to read config file: %w", err))
	}

	// load configuration into a config instance
	accessTokenDuration, err := time.ParseDuration(viper.Get("ACCESS_TOKEN_DURATION").(string))
	if err != nil {
		panic(fmt.Errorf("unable to read config file: %w", err))
	}
	refreshTokenDuration, err := time.ParseDuration(viper.Get("REFRESH_TOKEN_DURATION").(string))
	if err != nil {
		panic(fmt.Errorf("unable to read config file: %w", err))
	}
	return &config{
		databaseAddress:      viper.Get("DATABASE_ADDRESS").(string),
		testDatabaseAddress:  viper.Get("TEST_DATABASE_ADDRESS").(string),
		serverAddress:        viper.Get("SERVER_ADDRESS").(string),
		tokenSymmetricKey:    viper.Get("TOKEN_SYMMETRIC_KEY").(string),
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}
