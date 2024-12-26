package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4" // Добавляем импорт JWT
	"github.com/jmoiron/sqlx"
)

func AdminLogin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&credentials); err != nil {
			log.Printf("Ошибка привязки данных: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		var admin struct {
			AdminID  int    `db:"admin_id"`
			Username string `db:"username"`
			Email    string `db:"email"`
		}

		err := db.Get(&admin, `
			SELECT admin_id, username, email
			FROM administrators
			WHERE email = $1 AND password_hash = crypt($2, password_hash)
		`, credentials.Email, credentials.Password)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
			return
		} else if err != nil {
			log.Printf("Ошибка проверки администратора: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			return
		}

		// Генерируем JWT токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"admin_id": admin.AdminID,
			"email":    admin.Email,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // Токен действителен 24 часа
		})

		// Подписываем токен секретным ключом
		tokenString, err := token.SignedString([]byte("your-secret-key")) // В реальном приложении используйте безопасный ключ
		if err != nil {
			log.Printf("Ошибка создания токена: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Вход успешен",
			"token":   tokenString,
			"admin": gin.H{
				"id":       admin.AdminID,
				"username": admin.Username,
				"email":    admin.Email,
			},
		})
	}
}
