package api

import (
	"Chat-Server/api/ws"
	"Chat-Server/config"
	"Chat-Server/repository"
	"Chat-Server/token"
	"github.com/gin-contrib/cors"
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
	tokenMaker token.Maker
	configs    *config.Config
	chatHub    *ws.Hub
}

// NewServer initializes and returns a server
func NewServer(repository repository.Repository, tokenMaker token.Maker, configs *config.Config) *server {
	// get a gin router with default middlewares
	router := gin.Default()

	// CORS middleware configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}

	// use the CORS middleware with the custom configuration
	router.Use(cors.New(corsConfig))

	// create and return a server
	apiServer := server{
		repository: repository,
		router:     router,
		tokenMaker: tokenMaker,
		configs:    configs,
		chatHub:    ws.NewHub(),
	}

	// register custom validators
	registerCustomValidators()

	// add route handlers
	apiServer.addRouteHandlers()

	return &apiServer
}

// addRouteHandlers adds route handlers to server's router
func (s *server) addRouteHandlers() {
	s.router.POST("/api/signup", s.signup)
	s.router.POST("/api/login", s.login)
	s.router.POST("/api/refresh", s.refreshToken)

	// Set up static files using the Static method
	s.router.Static("/signup", "./static/signup")
	s.router.Static("/login", "./static/login")
	s.router.Static("/chat", "./static/chat")

	authGroup := s.router.Group("/", authMiddleware(s.tokenMaker))
	authGroup.GET("/api/chat", authMiddleware(s.tokenMaker), s.chat)

	// Handle requests that don't match any defined routes
	s.router.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/login")
	})
}

// Start starts server on the given address
func (s *server) Start(address string) error {
	go s.chatHub.RunChatHub(s.repository)
	return s.router.Run(address)
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
