package models

import "time"

// Модель для чата
type Chat struct {
	ChatID    int       `db:"chat_id" json:"chat_id"`
	UserID    int       `db:"user_id" json:"user_id"`
	AdminID   *int      `db:"admin_id" json:"admin_id"` // Используем указатель для nullable поля
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Модель для сообщений
type Message struct {
	MessageID  int       `db:"message_id" json:"message_id"`
	ChatID     int       `db:"chat_id" json:"chat_id"`
	SenderType string    `db:"sender_type" json:"sender_type"` // user или admin
	SenderID   int       `db:"sender_id" json:"sender_id"`
	Content    string    `db:"content" json:"content"`
	SentAt     time.Time `db:"sent_at" json:"sent_at"`
}
