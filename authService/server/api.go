package server

import (
	"net/http"
	"time"

	"authService/db"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db        *db.Database
	router    *gin.Engine
	jwtSecret string
}

func New(database *db.Database, jwtSecret string) *Server {
	router := gin.Default()

	s := &Server{
		db:        database,
		router:    router,
		jwtSecret: jwtSecret,
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
		auth.POST("/register", Register(s.db))
		auth.POST("/login", Login(s.db, s.jwtSecret))
		auth.GET("user/:id", GetUser(s.db))
		auth.DELETE("user/:id", DeleteUser(s.db))
	}

}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
