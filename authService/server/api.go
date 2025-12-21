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
		auth.POST("/register", Register(s.db))
		auth.POST("/login", Login(s.db))
	}
	users := s.router.Group("/users")
	{
		users.GET("/:id", GetUser(s.db))
		users.DELETE("/:id", DeleteUser(s.db))
		users.GET("/", ListUsers(s.db))
	}

}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
