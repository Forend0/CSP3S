package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Получение списка чатов
func GetChats(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var chats []struct {
			ChatID   int    `db:"chat_id" json:"chat_id"`
			UserName string `db:"username" json:"user_name"`
			UserID   int    `db:"user_id" json:"user_id"`
		}

		err := db.Select(&chats, `
			SELECT c.chat_id, u.username, u.user_id
			FROM chats c
			JOIN users u ON c.user_id = u.user_id
		`)
		if err != nil {
			log.Printf("Ошибка получения списка чатов: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка чатов"})
			return
		}

		c.JSON(http.StatusOK, chats)
	}
}

// Получение сообщений из чата
func GetMessages(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("chatId")
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			log.Printf("Некорректный chatId: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор чата"})
			return
		}

		var messages []struct {
			MessageID  int       `db:"message_id" json:"message_id"`
			ChatID     int       `db:"chat_id" json:"chat_id"`
			SenderType string    `db:"sender_type" json:"sender_type"`
			SenderID   int       `db:"sender_id" json:"sender_id"`
			Content    string    `db:"content" json:"content"`
			SentAt     time.Time `db:"sent_at" json:"sent_at"`
			IsRead     bool      `db:"is_read" json:"is_read"`
		}

		err = db.Select(&messages, `
            SELECT message_id, chat_id, sender_type, sender_id, content, sent_at, is_read
            FROM messages
            WHERE chat_id = $1
            ORDER BY sent_at ASC
        `, chatID)

		if err != nil {
			log.Printf("Ошибка получения сообщений: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения сообщений"})
			return
		}

		c.JSON(http.StatusOK, messages)
	}
}

// Отправка сообщения
func SendMessage(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("chatId")
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			log.Printf("Некорректный chatId: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор чата"})
			return
		}

		var message struct {
			SenderType string `json:"sender_type"` // user или admin
			SenderID   int    `json:"sender_id"`
			Content    string `json:"content"`
		}
		if err := c.ShouldBindJSON(&message); err != nil {
			log.Printf("Ошибка привязки данных сообщения: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные сообщения"})
			return
		}

		_, err = db.Exec(`
			INSERT INTO messages (chat_id, sender_type, sender_id, content)
			VALUES ($1, $2, $3, $4)
		`, chatID, message.SenderType, message.SenderID, message.Content)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки сообщения"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Сообщение отправлено"})
	}
}

// Создание чата
func CreateChat(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var chat struct {
			UserID  int `json:"user_id"`
			AdminID int `json:"admin_id"`
		}
		if err := c.ShouldBindJSON(&chat); err != nil {
			log.Printf("Ошибка привязки данных чата: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные чата"})
			return
		}

		// Проверяем существование администратора
		var exists bool
		err := db.Get(&exists, `
			SELECT EXISTS(SELECT 1 FROM administrators WHERE admin_id = $1)
		`, chat.AdminID)
		if err != nil || !exists {
			log.Printf("Администратор не найден: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Администратор не найден"})
			return
		}

		var chatID int
		err = db.Get(&chatID, `
			INSERT INTO chats (user_id, admin_id)
			VALUES ($1, $2)
			RETURNING chat_id
		`, chat.UserID, chat.AdminID)
		if err != nil {
			log.Printf("Ошибка создания чата: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания чата"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"chat_id": chatID})
	}
}

// GetUserChats возвращает чаты конкретного пользователя
func GetUserChats(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")

		var chats []struct {
			ChatID    int       `db:"chat_id" json:"chat_id"`
			UserID    int       `db:"user_id" json:"user_id"`
			AdminID   *int      `db:"admin_id" json:"admin_id"` // Используем указатель
			CreatedAt time.Time `db:"created_at" json:"created_at"`
		}

		err := db.Select(&chats, `
            SELECT chat_id, user_id, admin_id, created_at
            FROM chats
            WHERE user_id = $1
        `, userID)
		if err != nil {
			log.Printf("Ошибка получения чатов пользователя: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения чатов пользователя"})
			return
		}

		// Если чатов нет, возвращаем пустой массив
		if chats == nil {
			chats = make([]struct {
				ChatID    int       `db:"chat_id" json:"chat_id"`
				UserID    int       `db:"user_id" json:"user_id"`
				AdminID   *int      `db:"admin_id" json:"admin_id"`
				CreatedAt time.Time `db:"created_at" json:"created_at"`
			}, 0)
		}

		c.JSON(http.StatusOK, chats)
	}
}
