package handlers

import (
	"log"
	"net/http"
	"shopApi/internal/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func GetProducts(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		err := db.Select(&products, "SELECT * FROM products")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Ошибка получения списка продуктов",
				"details": err.Error(), // Логирование деталей ошибки
			})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

func GetProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := strings.TrimSpace(c.Param("id"))
		if idStr == "" {
			log.Println("Параметр id пуст")
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID пользователя отсутствует"})
			return
		}
		log.Printf("Полученный параметр idStr: '%s'", idStr)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID продукта"})
			return
		}
		var product models.Product
		err = db.Get(&product, "SELECT * FROM products WHERE product_id = $1", id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

func CreateProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product models.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		query := `INSERT INTO products (name, description, price, stock, image_url) 
                  VALUES (:name, :description, :price, :stock, :image_url) RETURNING product_id`

		rows, err := db.NamedQuery(query, &product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления продукта"})
			return
		}
		if rows.Next() {
			rows.Scan(&product.ProductID) // Присваиваем ID нового продукта
		}
		rows.Close()

		c.JSON(http.StatusCreated, product)
	}
}

func UpdateProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID продукта"})
			return
		}

		var product models.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			log.Printf("Ошибка привязки JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные"})
			return
		}

		log.Printf("Полученные данные для обновления: %+v", product)

		if product.Name == "" || product.Price == 0 || product.Stock == 0 {
			log.Printf("Обязательные поля отсутствуют: Name=%v, Price=%v, Stock=%v", product.Name, product.Price, product.Stock)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Обязательные поля не могут быть пустыми"})
			return
		}

		product.ProductID = id
		query := `UPDATE products SET name = :name, description = :description, price = :price, 
                  stock = :stock, image_url = :image_url WHERE product_id = :product_id`

		_, err = db.NamedExec(query, &product)
		if err != nil {
			log.Printf("Ошибка при обновлении продукта в БД: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления продукта"})
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

func DeleteProduct(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID продукта"})
			return
		}

		_, err = db.Exec("DELETE FROM products WHERE product_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления продукта"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Продукт успешно удален"})
	}
}
