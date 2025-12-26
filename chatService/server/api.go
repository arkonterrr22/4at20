package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"chatService/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	type Claims struct {
		UserID   string   `json:"user_id"`
		Username string   `json:"username"`
		Groups   []string `json:"groups"`
		jwt.RegisteredClaims
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[len(parts)-1], claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("groups", claims.Groups)
		c.Set("claims", claims)

		c.Next()
	}
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "chat",
			"time":    time.Now().UTC(),
		})
	})

	chat := s.router.Group("/chat")
	chat.Use(AuthMiddleware(s.jwtSecret))
	{
		chat.GET("/list", ListChats(s.db))
		chat.POST("/create", CreateChat(s.db))
		chat.GET("/:chatid/members/", GetMembers(s.db))
		chat.POST("/:chatid/members/add", AddMembers(s.db))
		chat.DELETE("/:chatid/members/remove", RemoveMembers(s.db))
		chat.DELETE("/:chatid/remove", DeleteChat(s.db))
		chat.GET("/:chatid/info", GetChatInfo(s.db))
		chat.PATCH("/:chatid/edit", EditChatInfo(s.db))
		chat.POST("/:chatid/messages/send", AddMesage(s.db))
		chat.GET("/:chatid/messages", GetMessages(s.db))
		chat.DELETE("/:chatid/messages/remove", DeleteMessages(s.db))
		chat.PATCH("/:chatid/messages/edit", EditMessage(s.db))
	}

}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
