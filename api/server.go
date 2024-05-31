package api

import (
	"Chat-Server/repository"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

// server represents a server
type server struct {
	router     *gin.Engine
	repository repository.Repository
}

// NewServer initializes and returns a server
func NewServer(repository repository.Repository) *server {
	// get a gin router with default middlewares
	router := gin.Default()

	// create and return a server
	apiServer := server{
		repository: repository,
		router:     router,
	}

	// register custom validators
	registerCustomValidators()

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

// registerCustomValidators registers custom validators to gin's binding package
func registerCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("validUsername", ValidUsername); err != nil {
			log.Fatal("could not register validUsername validator")
		}
		if err := v.RegisterValidation("validPassword", ValidPassword); err != nil {
			log.Fatal("could not register validPassword validator")
		}
	}
}
