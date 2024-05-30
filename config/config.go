package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// config holds configuration variables
type config struct {
	databaseAddress     string // address of the database
	testDatabaseAddress string // address of the test database
	serverAddress       string // address of the server
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
	return &config{
		databaseAddress:     viper.Get("DATABASE_ADDRESS").(string),
		testDatabaseAddress: viper.Get("TEST_DATABASE_ADDRESS").(string),
		serverAddress:       viper.Get("SERVER_ADDRESS").(string),
	}
}
