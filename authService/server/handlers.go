package server

import (
	"authService/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
			Username string `json:"username" binding:"required,max=50"`
			Login    string `json:"login" binding:"required,max=255"`
			Password string `json:"password" binding:"required,max=255"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var exists bool
		err := db.Pool.QueryRow(c.Request.Context(),
			"SELECT EXISTS(SELECT 1 FROM auth WHERE login = $1 AND password = $2)",
			req.Login, req.Password,
		).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}

		userID := uuid.New()
		tx, err := db.Pool.Begin(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction error"})
			return
		}
		defer tx.Rollback(c.Request.Context())

		_, err = tx.Exec(c.Request.Context(),
			"INSERT INTO users (id, username) VALUES ($1, $2)",
			userID, req.Username,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании пользователя"})
			return
		}
		_, err = tx.Exec(c.Request.Context(),
			"INSERT INTO auth (user_id, login, password) VALUES ($1, $2, $3)",
			userID, req.Login, req.Password,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении пароля"})
			return
		}
		_, _ = tx.Exec(c.Request.Context(),
			`INSERT INTO user_group (user_id, group_id) 
			 VALUES ($1, '00000000-0000-0000-0000-000000000000')
			 ON CONFLICT DO NOTHING`,
			userID,
		)

		if err := tx.Commit(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Коммит не осуществлён"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Регистрация прошла успешно",
			"user_id":  userID.String(),
			"username": req.Username,
		})
	}
}

func Login(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Login    string `json:"login" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userID uuid.UUID
		var password string
		var username string
		err := db.Pool.QueryRow(c.Request.Context(),
			`SELECT a.user_id, a.password, u.username
			 FROM auth a
			 JOIN users u ON a.user_id = u.id
			 WHERE a.login = $1`,
			req.Login,
		).Scan(&userID, &password, &username)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин"})
			return
		}

		if password != req.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
			return
		}

		rows, err := db.Pool.Query(c.Request.Context(),
			`SELECT ug.group_id
			 FROM user_group ug
			 WHERE ug.user_id = $1`,
			userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении групп"})
			return
		}
		defer rows.Close()
		var groups []uuid.UUID
		for rows.Next() {
			var groupID uuid.UUID
			if err := rows.Scan(&groupID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки групп"})
				return
			}
			groups = append(groups, groupID)
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при чтении групп"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  userID.String(),
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

		tokenString, err := token.SignedString([]byte("super-secret-key"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"user_id":  userID.String(),
			"username": username,
			"groups":   groups,
			"token":    tokenString,
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

		c.JSON(http.StatusOK, gin.H{
			"message": "Пользователь удалён",
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
