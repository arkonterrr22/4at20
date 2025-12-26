package server

import (
	"chatService/db"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Chat struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Pic        string    `json:"pic"`
	Created_at time.Time `json:"created_at"`
}

type Message struct {
	Id         int64     `json:"id"`
	Chat_ID    uuid.UUID `json:"chat_id"`
	User_ID    uuid.UUID `json:"user_id"`
	Text       string    `json:"text"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
}

func ListChats(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		rows, err := db.Pool.Query(c, `
			SELECT c.*
			FROM chats c
			JOIN user_chat uc ON uc.chat_id = c.id
			WHERE uc.user_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()
		chats := make([]Chat, 0)
		for rows.Next() {
			var chat Chat
			if err := rows.Scan(&chat.Id, &chat.Name, &chat.Pic, &chat.Created_at); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			chats = append(chats, chat)
		}

		c.JSON(http.StatusOK, chats)
	}
}

func CreateChat(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string  `json:"name" binding:"required"`
			Pic  *string `json:"pic"`
		}
		userID := c.GetString("user_id")
		fmt.Printf("userid: %s", userID)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var chat Chat
		err := db.Pool.QueryRow(
			c.Request.Context(),
			`INSERT INTO chats (name, pic)
			VALUES ($1, $2)
			RETURNING *`,
			req.Name, req.Pic,
		).Scan(
			&chat.Id,
			&chat.Name,
			&chat.Pic,
			&chat.Created_at,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cmdTag, err := db.Pool.Exec(
			c.Request.Context(),
			`INSERT INTO user_chat (user_id, chat_id)
			VALUES ($1, $2)
			`, c.GetString("user_id"), chat.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if cmdTag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "couldnt connect user with chat"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"chat": chat, "userID": userID})
	}
}

func DeleteChat(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		cmdTag, err := db.Pool.Exec(
			c.Request.Context(),
			`DELETE FROM chats WHERE id = $1`,
			chatID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if cmdTag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
			return
		}

		c.Status(http.StatusOK)
	}
}

func GetMembers(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		rows, err := db.Pool.Query(c, `
			SELECT uc.user_id
			FROM user_chat uc
			WHERE uc.chat_id = $1
		`, chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()
		members := make([]uuid.UUID, 0)
		for rows.Next() {
			var member uuid.UUID
			if err := rows.Scan(&member); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			members = append(members, member)
		}

		c.JSON(http.StatusOK, gin.H{"members": members})
	}
}

func AddMembers(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Members []uuid.UUID `json:"members" binding:"required"`
		}
		chatID := c.Param("chatid")

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmdTag, err := db.Pool.Exec(c.Request.Context(),
			"INSERT INTO user_chat (chat_id, user_id) SELECT $1, unnest($2::uuid[])",
			chatID, req.Members,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при связывании пользователей с чатом: %s", err)})
			return
		}

		c.JSON(http.StatusOK, cmdTag.RowsAffected())
	}
}

func RemoveMembers(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Members []uuid.UUID `json:"members" binding:"required"`
		}
		chatID := c.Param("chatid")

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmdTag, err := db.Pool.Exec(c.Request.Context(),
			`DELETE FROM user_chat
			WHERE chat_id = $1
			AND user_id = ANY($2::uuid[])`,
			chatID, req.Members,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при связывании пользователей с чатом: %s", err)})
			return
		}

		c.JSON(http.StatusOK, cmdTag.RowsAffected())
	}
}

func GetChatInfo(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		var chat Chat
		err := db.Pool.QueryRow(c, `
			SELECT *
			FROM chats c
			WHERE c.id = $1
		`, chatID).Scan(&chat.Id, &chat.Name, &chat.Pic, &chat.Created_at)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"chat": chat})
	}
}

func EditChatInfo(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		var chat Chat
		var req struct {
			Name *string `json:"name"`
			Pic  *string `json:"pic"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		fmt.Printf("name: %s, pic: %s", *req.Name, *req.Pic)
		err := db.Pool.QueryRow(c, `
			UPDATE chats c
			SET
				name = COALESCE($1, name),
				pic = COALESCE($2, pic)
			WHERE c.id = $3
			RETURNING c.*
		`, *req.Name, *req.Pic, chatID).Scan(&chat.Id, &chat.Name, &chat.Pic, &chat.Created_at)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"chat": chat})
	}
}

func AddMesage(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		userID := c.GetString(("user_id"))
		var req struct {
			Text    string `json:"text" binding:"required"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var msg Message
		err := db.Pool.QueryRow(c, `
			INSERT INTO messages (chat_id, user_id, text, content)
			VALUES ($1, $2, $3, $4)
			RETURNING *`, chatID, userID, req.Text, req.Content).Scan(&msg.Id, &msg.Chat_ID, &msg.User_ID, &msg.Text, &msg.Content, &msg.Created_at)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}

func GetMessages(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		page := c.Query("page")
		limit := c.Query("limit")
		rows, err := db.Pool.Query(c, `
			SELECT *
			FROM messages m
			WHERE m.chat_id = $1
			ORDER BY m.created_at desc
			LIMIT $2
			OFFSET $3;
		`, chatID, limit, page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error1": err.Error(),
			})
			return
		}
		defer rows.Close()
		messages := make([]Message, 0)
		for rows.Next() {
			var msg Message
			if err := rows.Scan(&msg.Id, &msg.Chat_ID, &msg.User_ID, &msg.Text, &msg.Content, &msg.Created_at); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error2": err.Error(),
				})
				return
			}
			messages = append(messages, msg)
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}

func DeleteMessages(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		// userID := c.GetString(("user_id"))
		var req struct {
			Messages []int64 `json:"messages" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		cmdTag, err := db.Pool.Exec(c.Request.Context(),
			`DELETE FROM messages
			WHERE chat_id = $1
			AND id = ANY($2)`,
			chatID, req.Messages,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при удалении сообщений: %s", err)})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Удалено сообщений: %d", cmdTag.RowsAffected())})
	}
}

func EditMessage(db *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatID := c.Param("chatid")
		msgID := c.Query("id")
		var msg Message
		var req struct {
			Text    *string `json:"text"`
			Content *string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := db.Pool.QueryRow(c, `
			UPDATE messages m
			SET
				text = COALESCE($1, text),
				content = COALESCE($2, content)
			WHERE m.id = $3 AND m.chat_id = $4
			RETURNING m.*
		`, *req.Text, *req.Content, msgID, chatID).Scan(&msg.Id, &msg.Chat_ID, &msg.User_ID, &msg.Text, &msg.Content, &msg.Created_at)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"message": msg})
	}
}
