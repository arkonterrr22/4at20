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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
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
			"service": "auth",
			"time":    time.Now().UTC(),
		})
	})

	chat := s.router.Group("/chat")
	chat.Use(AuthMiddleware(s.jwtSecret))
	{
		chat.GET("/list", ListChats(s.db))
		chat.POST("/create", CreateChat(s.db))
		chat.GET("/:chatid/info", GetChatInfo(s.db))
		chat.GET("/:chatid/messages", GetChatMessages(s.db))
		chat.DELETE("/:chatid/messages", DeleteMessages(s.db))
		chat.POST("/:chatid/messages/send", SendMesage(s.db))
		chat.GET("/:chatid/messages/:msgid", GetMessage(s.db))
		chat.PATCH("/:chatid/messages/:msgid", EditMessage(s.db))
	}

}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}
