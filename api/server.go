package api

import (
	"Chat-Server/repository"
	"github.com/gin-gonic/gin"
)

// server represents a server
type server struct {
	router     *gin.Engine
	repository repository.Repository
}

// NewServer initializes and returns a server
func NewServer() *server {
	// get a gin router with default middlewares
	router := gin.Default()

	// get a repository to interact with a postgresql database
	repository := repository.NewRepository(repository.Postgres)

	// create and return a server
	return &server{
		repository: repository,
		router:     router,
	}
}
