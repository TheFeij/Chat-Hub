package api

import (
	"Chat-Server/repository"
	"github.com/gin-gonic/gin"
	"net/http"
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
	apiServer := server{
		repository: repository,
		router:     router,
	}

	// add route handlers
	apiServer.addRouteHandlers()

	return &apiServer
}

// addRouteHandlers adds route handlers to server's router
func (s *server) addRouteHandlers() {
	s.router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "Welcome!")
	})

	// TODO: add other route handlers
}
