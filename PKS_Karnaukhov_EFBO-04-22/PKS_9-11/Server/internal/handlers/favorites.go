package handlers

import (
	"log"
	"net/http"
	"shopApi/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// handlers/favorites.go
func GetFavorites(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования userID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		var favorites []models.Favorite
		err = db.Select(&favorites, `
            SELECT favorite_id, user_id, product_id, created_at 
            FROM favorites 
            WHERE user_id = $1
            ORDER BY created_at DESC
        `, userID)

		if err != nil {
			log.Printf("Ошибка запроса к базе данных: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения избранного"})
			return
		}

		// Всегда возвращаем массив, даже если он пустой
		if favorites == nil {
			favorites = make([]models.Favorite, 0)
		}

		c.JSON(http.StatusOK, favorites)
	}
}

func AddToFavorites(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("userId")
		var item struct {
			ProductID int `json:"product_id"`
		}
		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}
		_, err := db.Exec("INSERT INTO Favorites (user_id, product_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			userId, item.ProductID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления в избранное"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в избранное"})
	}
}

func RemoveFromFavorites(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("userId")
		productId := c.Param("productId")

		// Логирование параметров
		log.Println("Received userId:", userId, "productId:", productId)

		_, err := db.Exec("DELETE FROM Favorites WHERE user_id = $1 AND product_id = $2", userId, productId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления из избранного"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Товар удален из избранного"})
	}
}
