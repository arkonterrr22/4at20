package server

import (
	"net/http"
	"time"

	"authService/db"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db     *db.Database
	router *gin.Engine
}

func New(database *db.Database) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	s := &Server{
		db:     database,
		router: router,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "auth",
			"time":    time.Now().UTC(),
		})
	})

	auth := s.router.Group("/auth")
	{
		auth.GET("/hi", Hi(s.db))
		// Auth
		// auth.POST("/register", Register(s.db))
		// auth.POST("/login", Login(s.db))

		// Users
		// v1.GET("/users/:id", handlers.GetUser(s.db))
		// v1.DELETE("/users/:id", handlers.DeleteUser(s.db))
		// v1.GET("/users", handlers.ListUsers(s.db))
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
