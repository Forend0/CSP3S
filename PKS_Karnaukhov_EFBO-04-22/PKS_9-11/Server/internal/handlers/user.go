package handlers

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"shopApi/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// CreateUser создает нового пользователя
func CreateUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			log.Println("Ошибка при биндинге данных:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		// Логирование данных, переданных на сервер
		log.Printf("Регистрация пользователя: username=%s, email=%s, password=%s", user.UserName, user.Email, user.PasswordHash)

		// Проверка, существует ли уже пользователь с таким email
		var existingUser models.User
		err := db.Get(&existingUser, "SELECT * FROM users WHERE email = $1", user.Email)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже существует"})
			return
		}

		// Генерация уникального user_id
		user.UserID = rand.Intn(1000000) // Вместо UUID используем случайный ID
		user.CreatedAt = time.Now()

		// Вставка пользователя в базу данных
		_, err = db.Exec(
			"INSERT INTO users (user_id, username, email, password_hash, created_at) VALUES ($1, $2, $3, $4, $5)",
			user.UserID, user.UserName, user.Email, user.PasswordHash, user.CreatedAt,
		)
		if err != nil {
			log.Println("Ошибка добавления пользователя в базу данных:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Пользователь успешно создан",
			"user_id": user.UserID,
		})
	}
}

// LoginUser авторизует пользователя
func LoginUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&credentials); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		log.Printf("Получены данные для авторизации: email=%s, password=%s", credentials.Email, credentials.Password)

		var user models.User
		err := db.Get(&user, "SELECT * FROM users WHERE email = $1", credentials.Email)
		if err != nil {
			log.Println("Пользователь не найден:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
			return
		}

		log.Printf("Данные пользователя из базы данных: email=%s, password=%s", user.Email, user.PasswordHash)

		// Проверка пароля (сравнение открытого текста)
		if user.PasswordHash != credentials.Password {
			log.Printf("Пароль не совпадает! Введённый пароль: %s, Пароль из базы: %s", credentials.Password, user.PasswordHash)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":    user.UserID,
			"username":   user.UserName,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		})
	}
}

func RegisterUser(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
			return
		}

		// Проверяем, существует ли пользователь с таким email
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		if exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
			return
		}

		// Создаем нового пользователя
		var userID int
		err = db.QueryRow(`
            INSERT INTO users (username, email, password_hash)
            VALUES ($1, $2, $3)
            RETURNING user_id`,
			user.Username, user.Email, user.Password).Scan(&userID)

		if err != nil {
			log.Printf("Error creating user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
			return
		}

		log.Printf("Created new user: ID=%d, Username=%s, Email=%s", userID, user.Username, user.Email)

		c.JSON(http.StatusCreated, gin.H{
			"user_id":    userID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": time.Now(),
		})
	}
}

// GetUserInfo возвращает информацию о текущем пользователе
func GetUserInfo(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userId")

		var user models.User
		err := db.Get(&user, "SELECT user_id, username, email, created_at FROM users WHERE user_id = $1", userID)
		if err != nil {
			log.Println("Ошибка получения информации о пользователе:", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
