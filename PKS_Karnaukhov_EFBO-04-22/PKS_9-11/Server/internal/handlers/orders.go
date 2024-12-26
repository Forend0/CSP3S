package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"shopApi/internal/models"
	"strconv"
	"strings"
	"time" // Импортируем пакет time для работы с датами

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func GetOrders(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		log.Println("Полученный параметр idStr:", idStr)

		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			log.Println("Ошибка преобразования idStr в int:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
			return
		}

		// Получаем заказы с явным указанием колонок
		var orders []struct {
			OrderID   int       `db:"order_id"`
			UserID    int       `db:"user_id"`
			Total     float64   `db:"total"`
			Status    string    `db:"status"`
			CreatedAt time.Time `db:"created_at"`
		}

		err = db.Select(&orders, `
            SELECT order_id, user_id, total, status, created_at 
            FROM orders 
            WHERE user_id = $1
        `, id)
		if err != nil {
			log.Println("Ошибка запроса к базе данных:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заказов"})
			return
		}

		// Формируем результат
		var result []models.Order
		for _, o := range orders {
			// Получаем продукты для каждого заказа
			var dbProducts []struct {
				ProductID int `db:"product_id"`
				Quantity  int `db:"quantity"`
			}
			err = db.Select(&dbProducts, `
                SELECT product_id, quantity 
                FROM order_products 
                WHERE order_id = $1
            `, o.OrderID)
			if err != nil {
				log.Printf("Ошибка получения продуктов для заказа %d: %v", o.OrderID, err)
				continue
			}

			// Преобразуем продукты в нужный формат
			var products []struct {
				ProductID int `json:"product_id"`
				Quantity  int `json:"quantity"`
			}
			for _, p := range dbProducts {
				products = append(products, struct {
					ProductID int `json:"product_id"`
					Quantity  int `json:"quantity"`
				}{
					ProductID: p.ProductID,
					Quantity:  p.Quantity,
				})
			}

			// Добавляем заказ в результат
			result = append(result, models.Order{
				OrderID:   o.OrderID,
				UserID:    o.UserID,
				Total:     o.Total,
				Status:    o.Status,
				CreatedAt: o.CreatedAt,
				Products:  products,
			})
		}

		c.JSON(http.StatusOK, result)
	}
}
func CreateOrder(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order

		// Читаем тело запроса в буфер
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Ошибка чтения тела запроса: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка чтения запроса"})
			return
		}

		// Логируем тело запроса
		log.Printf("Полученное тело запроса: %s", string(body))

		// Создаем новый reader из буфера для ShouldBindJSON
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Привязка данных к структуре Order
		if err := c.ShouldBindJSON(&order); err != nil {
			log.Printf("Ошибка биндинга данных: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		log.Printf("Привязанные данные заказа: %+v", order)

		// Проверяем, что массив продуктов не пустой
		if len(order.Products) == 0 {
			log.Println("Ошибка: заказ не содержит продуктов")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Заказ должен содержать хотя бы один продукт"})
			return
		}

		// Начинаем транзакцию
		tx, err := db.Beginx()
		if err != nil {
			log.Printf("Ошибка инициализации транзакции: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка инициализации транзакции"})
			return
		}

		// Сначала создаем заказ и получаем его ID
		var orderID int
		err = tx.QueryRow(`
    INSERT INTO orders (status, user_id, created_at, total) 
    VALUES ($1, $2, $3, $4) 
    RETURNING order_id
`, order.Status, order.UserID, order.CreatedAt, order.Total).Scan(&orderID)

		if err != nil {
			tx.Rollback()
			log.Printf("Ошибка создания заказа: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания заказа"})
			return
		}

		// Обновляем orderID
		order.OrderID = orderID

		// Теперь добавляем продукты
		for _, product := range order.Products {
			_, err = tx.Exec(`
                INSERT INTO order_products (order_id, product_id, quantity)
                VALUES ($1, $2, $3)
            `, orderID, product.ProductID, product.Quantity)
			if err != nil {
				tx.Rollback()
				log.Printf("Ошибка добавления продукта в заказ: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления продукта в заказ"})
				return
			}
		}

		// Завершаем транзакцию
		if err := tx.Commit(); err != nil {
			log.Printf("Ошибка завершения транзакции: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка завершения заказа"})
			return
		}

		// Отправляем успешный ответ
		log.Printf("Заказ создан успешно: %+v", order)
		c.JSON(http.StatusCreated, order)
	}
}

func DeleteOrder(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		log.Println("Полученный параметр idStr:", idStr)

		orderID, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			log.Println("Ошибка преобразования idStr в int:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID заказа"})
			return
		}

		// Удаляем записи о продуктах в заказе
		_, err = db.Exec(`
			DELETE FROM order_products
			WHERE order_id = $1
		`, orderID)
		if err != nil {
			log.Printf("Ошибка удаления продуктов для заказа %d: %v", orderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления продуктов из заказа"})
			return
		}

		// Удаляем сам заказ
		_, err = db.Exec(`
			DELETE FROM orders
			WHERE order_id = $1
		`, orderID)
		if err != nil {
			log.Printf("Ошибка удаления заказа %d: %v", orderID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления заказа"})
			return
		}

		log.Printf("Заказ %d успешно удален", orderID)
		c.JSON(http.StatusOK, gin.H{"message": "Заказ успешно удален"})
	}
}
