package api

// Path: github.com/pynezz/internal/api/http.go

// To simplify, we'll use Gin as the HTTP framework for the API server.
// This will allow us to quickly set up the server and focus on the actual functionality of the application.
// We'll start by adding the necessary dependencies to the go.mod file and then create the main server struct.

import (
	"github.com/gin-gonic/gin"
	"github.com/pynezz/bivrost/internal/config"
)

// Server is the main struct for the API server
type Server struct {
	config *config.Config
	router *gin.Engine
}

// NewServer creates a new API server with the given configuration
func NewServer(cfg *config.Config) *Server {
	// Create a new server with the given configuration
	return &Server{
		config: cfg,
		router: gin.Default(),
	}
}

// Start starts the API server
func (s *Server) Start() {
	// Start the API server
	s.router.Run(":8080")
}
