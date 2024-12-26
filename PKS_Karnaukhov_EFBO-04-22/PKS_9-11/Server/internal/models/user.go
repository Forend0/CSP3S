package models

import "time"

// User представляет структуру пользователя
type User struct {
	UserID       int       `json:"user_id" db:"user_id"`
	UserName     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"password" db:"password_hash"` // Храним пароль в открытом виде
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
