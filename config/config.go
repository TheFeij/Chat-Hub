package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// Config holds configuration variables
type Config struct {
	databaseAddress        string        // address of the database
	testDatabaseAddress    string        // address of the test database
	serverAddress          string        // address of the server
	tokenSymmetricKey      string        // symmetric key of to make and verify tokens
	accessTokenDuration    time.Duration // access token duration
	refreshTokenDuration   time.Duration // refresh token duration
	accessTokenCookiePath  string        // access token's cookie path
	refreshTokenCookiePath string        // refresh token's cookie path
	usernameCookiePath     string        // username's cookie path
}

// AccessTokenCookiePath returns the access token cookie path
func (c Config) AccessTokenCookiePath() string {
	return c.accessTokenCookiePath
}

// RefreshTokenCookiePath returns the refresh token cookie path
func (c Config) RefreshTokenCookiePath() string {
	return c.refreshTokenCookiePath
}

// UsernameCookiePath returns username cookie path
func (c Config) UsernameCookiePath() string {
	return c.usernameCookiePath
}

// TokenSymmetricKey returns token symmetric key
func (c Config) TokenSymmetricKey() string {
	return c.tokenSymmetricKey
}

// AccessTokenDuration returns access token duration
func (c Config) AccessTokenDuration() time.Duration {
	return c.accessTokenDuration
}

// RefreshTokenDuration returns refresh token duration
func (c Config) RefreshTokenDuration() time.Duration {
	return c.refreshTokenDuration
}

// DatabaseAddress returns database address
func (c Config) DatabaseAddress() string {
	return c.databaseAddress
}

// TestDatabaseAddress returns test database address
func (c Config) TestDatabaseAddress() string {
	return c.testDatabaseAddress
}

// ServerAddress returns server address
func (c Config) ServerAddress() string {
	return c.serverAddress
}

// GetConfig returns a config object loaded with the config variables of the
// file specified in the input
func GetConfig(configFileName, configFileType, configFilePath string) *Config {
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
	return &Config{
		databaseAddress:        viper.Get("DATABASE_ADDRESS").(string),
		testDatabaseAddress:    viper.Get("TEST_DATABASE_ADDRESS").(string),
		serverAddress:          viper.Get("SERVER_ADDRESS").(string),
		tokenSymmetricKey:      viper.Get("TOKEN_SYMMETRIC_KEY").(string),
		accessTokenDuration:    accessTokenDuration,
		refreshTokenDuration:   refreshTokenDuration,
		accessTokenCookiePath:  viper.Get("ACCESS_TOKEN_COOKIE_PATH").(string),
		refreshTokenCookiePath: viper.Get("REFRESH_TOKEN_COOKIE_PATH").(string),
		usernameCookiePath:     viper.Get("USERNAME_COOKIE_PATH").(string),
	}
}
