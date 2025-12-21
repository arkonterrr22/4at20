package server

import (
	"net/http"
	"time"

	"authService/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Hi(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"reply": "hi",
		})
	}
}

func Register(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required,min=3,max=50"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Проверяем существование пользователя
		var exists bool
		err := db.Pool.QueryRow(c.Request.Context(),
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)",
			req.Username,
		).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}

		// Создаем пользователя
		userID := uuid.New()
		tx, err := db.Pool.Begin(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
			return
		}
		defer tx.Rollback(c.Request.Context())

		// 1. Вставляем пользователя
		_, err = tx.Exec(c.Request.Context(),
			"INSERT INTO users (id, username) VALUES ($1, $2)",
			userID, req.Username,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// 2. Сохраняем пароль (TODO: хешировать!)
		_, err = tx.Exec(c.Request.Context(),
			"INSERT INTO auth (user_id, password_hash) VALUES ($1, $2)",
			userID, req.Password, // ВРЕМЕННО - хешировать в продакшене!
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save password"})
			return
		}

		// 3. Добавляем в группу "Все пользователи" если есть
		_, _ = tx.Exec(c.Request.Context(),
			`INSERT INTO user_group (user_id, group_id) 
			 VALUES ($1, '00000000-0000-0000-0000-000000000000')
			 ON CONFLICT DO NOTHING`,
			userID,
		)

		if err := tx.Commit(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "User registered successfully",
			"user_id":  userID.String(),
			"username": req.Username,
		})
	}
}

func Login(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userID uuid.UUID
		var passwordHash string
		err := db.Pool.QueryRow(c.Request.Context(),
			`SELECT u.id, a.password_hash 
			 FROM users u
			 JOIN auth a ON a.user_id = u.id
			 WHERE u.username = $1`,
			req.Username,
		).Scan(&userID, &passwordHash)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// TODO: Использовать bcrypt.CompareHashAndPassword()
		if passwordHash != req.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// TODO: Генерировать JWT токен
		token := "jwt-token-placeholder-" + userID.String()

		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"user_id":  userID.String(),
			"username": req.Username,
			"token":    token,
		})
	}
}

func GetUser(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var username string
		var createdAt time.Time
		err := db.Pool.QueryRow(c.Request.Context(),
			"SELECT username, created_at FROM users WHERE id = $1",
			userID,
		).Scan(&username, &createdAt)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         userID,
			"username":   username,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
}

func DeleteUser(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		_, err := db.Pool.Exec(c.Request.Context(),
			"DELETE FROM users WHERE id = $1",
			userID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		// TODO: Отправить событие в RabbitMQ о удалении пользователя

		c.JSON(http.StatusOK, gin.H{
			"message": "User deleted successfully",
			"user_id": userID,
		})
	}
}

func ListUsers(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "50")
		offset := c.DefaultQuery("offset", "0")

		rows, err := db.Pool.Query(c.Request.Context(),
			`SELECT id, username, created_at 
			 FROM users 
			 ORDER BY created_at DESC 
			 LIMIT $1 OFFSET $2`,
			limit, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer rows.Close()

		users := []gin.H{}
		for rows.Next() {
			var id uuid.UUID
			var username string
			var createdAt time.Time

			if err := rows.Scan(&id, &username, &createdAt); err != nil {
				continue
			}

			users = append(users, gin.H{
				"id":         id.String(),
				"username":   username,
				"created_at": createdAt.Format(time.RFC3339),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	}
}
