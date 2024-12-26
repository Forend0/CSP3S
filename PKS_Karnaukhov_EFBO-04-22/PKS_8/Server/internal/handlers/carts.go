package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func GetCart(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования userID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		// Получаем корзину с информацией о продуктах через JOIN
		query := `
            SELECT 
                c.cart_id,
                c.user_id,
                c.product_id,
                c.quantity,
                c.added_at,
                p.name,
                p.description,
                p.price,
                p.stock,
                p.image_url
            FROM cart c
            INNER JOIN products p ON c.product_id = p.product_id
            WHERE c.user_id = $1
            ORDER BY c.added_at DESC
        `

		type CartItemWithProduct struct {
			CartID      int       `json:"cart_id" db:"cart_id"`
			UserID      int       `json:"user_id" db:"user_id"`
			ProductID   int       `json:"product_id" db:"product_id"`
			Quantity    int       `json:"quantity" db:"quantity"`
			AddedAt     time.Time `json:"added_at" db:"added_at"`
			Name        string    `json:"name" db:"name"`
			Description string    `json:"description" db:"description"`
			Price       float64   `json:"price" db:"price"`
			Stock       int       `json:"stock" db:"stock"`
			ImageURL    string    `json:"image_url" db:"image_url"`
		}

		var items []CartItemWithProduct

		// Выполняем запрос и логируем детали
		log.Printf("Выполнение запроса для userID: %d", userID)
		err = db.Select(&items, query, userID)
		if err != nil {
			log.Printf("Ошибка запроса к БД: %v", err)
			log.Printf("SQL запрос: %s", query)
			log.Printf("Параметры: userID=%d", userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения корзины"})
			return
		}

		// Если корзина пуста, возвращаем пустой массив
		if items == nil {
			items = make([]CartItemWithProduct, 0)
		}

		// Логируем результат
		log.Printf("Получено %d товаров из корзины для пользователя %d", len(items), userID)
		for i, item := range items {
			log.Printf("Товар %d: ID=%d, Name=%s, Price=%.2f, Stock=%d",
				i+1, item.ProductID, item.Name, item.Price, item.Stock)
		}

		c.JSON(http.StatusOK, items)
	}
}

func AddToCart(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования userID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		var item struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		if err := c.ShouldBindJSON(&item); err != nil {
			log.Printf("Ошибка разбора JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		log.Printf("Добавление в корзину: userID=%d, productID=%d, quantity=%d",
			userID, item.ProductID, item.Quantity)

		// Используем CTE для вставки или обновления
		query := `
            WITH upsert AS (
                UPDATE cart 
                SET quantity = cart.quantity + $3
                WHERE user_id = $1 AND product_id = $2
                RETURNING *
            )
            INSERT INTO cart (user_id, product_id, quantity)
            SELECT $1, $2, $3
            WHERE NOT EXISTS (SELECT * FROM upsert);
        `

		_, err = db.Exec(query, userID, item.ProductID, item.Quantity)
		if err != nil {
			log.Printf("Ошибка добавления в корзину: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления в корзину"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в корзину"})
	}
}

// Удаление товара из корзины
func RemoveFromCart(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		productIDStr := c.Param("productId")

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования userID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		productID, err := strconv.Atoi(productIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования productID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID товара"})
			return
		}

		log.Printf("Удаление товара из корзины: userID=%d, productID=%d", userID, productID)

		// Удаляем товар из корзины
		result, err := db.Exec(`
            DELETE FROM cart 
            WHERE user_id = $1 AND product_id = $2`,
			userID, productID)

		if err != nil {
			log.Printf("Ошибка удаления из корзины: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления из корзины"})
			return
		}

		rows, err := result.RowsAffected()
		if err != nil || rows == 0 {
			log.Printf("Товар не найден в корзине: userID=%d, productID=%d", userID, productID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден в корзине"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Товар удален из корзины"})
	}
}

// Обновление количества товара
func UpdateCartQuantity(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Printf("Ошибка преобразования userID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		var body struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			log.Printf("Ошибка разбора JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		// Обновляем количество
		result, err := db.Exec(`
            UPDATE cart 
            SET quantity = quantity + $1
            WHERE user_id = $2 AND product_id = $3
            RETURNING quantity`,
			body.Quantity, userID, body.ProductID)

		if err != nil {
			log.Printf("Ошибка обновления количества: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления количества"})
			return
		}

		rows, err := result.RowsAffected()
		if err != nil || rows == 0 {
			log.Printf("Товар не найден в корзине: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден в корзине"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Количество обновлено"})
	}
}
func IncreaseQuantity(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		var body struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		// Проверяем существование записи
		var exists bool
		err = db.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM cart 
                WHERE user_id = $1 AND product_id = $2
            )`, userID, body.ProductID).Scan(&exists)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки существования"})
			return
		}

		if exists {
			// Обновляем существующую запись
			_, err = db.Exec(`
                UPDATE cart 
                SET quantity = quantity + $1 
                WHERE user_id = $2 AND product_id = $3`,
				body.Quantity, userID, body.ProductID)
		} else {
			// Создаем новую запись
			_, err = db.Exec(`
                INSERT INTO cart (user_id, product_id, quantity) 
                VALUES ($1, $2, $3)`,
				userID, body.ProductID, body.Quantity)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления количества"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Количество обновлено"})
	}
}

func DecreaseQuantity(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		var body struct {
			ProductID int `json:"product_id"`
			Quantity  int `json:"quantity"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		// Получаем текущее количество
		var currentQuantity int
		err = db.QueryRow(`
            SELECT quantity 
            FROM cart 
            WHERE user_id = $1 AND product_id = $2`,
			userID, body.ProductID).Scan(&currentQuantity)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения количества"})
			return
		}

		if currentQuantity <= body.Quantity {
			// Если количество станет 0 или меньше, удаляем товар
			_, err = db.Exec(`
                DELETE FROM cart 
                WHERE user_id = $1 AND product_id = $2`,
				userID, body.ProductID)
		} else {
			// Иначе уменьшаем количество
			_, err = db.Exec(`
                UPDATE cart 
                SET quantity = quantity - $1 
                WHERE user_id = $2 AND product_id = $3`,
				body.Quantity, userID, body.ProductID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления количества"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Количество обновлено"})
	}
}
