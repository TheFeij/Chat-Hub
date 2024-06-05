package main

import (
	"Chat-Server/api"
	"Chat-Server/config"
	"Chat-Server/repository/db/postgres"
	"Chat-Server/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	// load configurations and env variables
	configs := config.GetConfig("config", "json", "./config")

	// enable pretty logging for development
	if !configs.IsProductionEnv() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// connect to repository
	repository := postgres.GetPostgresRepository(configs.DatabaseAddress())

	// get a new paseto token maker
	tokenMaker, err := token.NewPasetoMaker(configs.TokenSymmetricKey())
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create a new paseto token maker")
	}

	// get a new server instance
	server := api.NewServer(repository, tokenMaker, configs)

	// start server
	err = server.Start(configs.ServerAddress())
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
