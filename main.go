package main

import (
	"Chat-Server/api"
	"Chat-Server/config"
	"Chat-Server/repository/db/postgres"
	"Chat-Server/token"
)

func main() {
	configs := config.GetConfig("config", "json", "./config")

	repository := postgres.GetPostgresRepository(configs.DatabaseAddress())

	tokenMaker, _ := token.NewPasetoMaker(configs.TokenSymmetricKey())

	server := api.NewServer(repository, tokenMaker, configs)

	server.Start(configs.ServerAddress())
}
