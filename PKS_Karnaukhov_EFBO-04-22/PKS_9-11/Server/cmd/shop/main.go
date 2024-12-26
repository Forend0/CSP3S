package main

import (
	"log"
	"shopApi/internal/handlers"
	"shopApi/pkg/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "shopApi/docs"
)

func main() {
	// Подключаемся к базе данных
	db, err := db.ConnectDB()
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	router := gin.Default()

	// Настройка CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Разрешаем запросы с этого источника (из фронтенда)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Swagger UI по адресу /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Роуты для продуктов
	router.GET("/products", handlers.GetProducts(db))
	router.GET("/products/:id", handlers.GetProduct(db))
	router.POST("/products", handlers.CreateProduct(db))
	router.PUT("/products/:id", handlers.UpdateProduct(db))
	router.DELETE("/products/:id", handlers.DeleteProduct(db))

	// Роуты для корзины
	router.POST("/cart/:userId/increment", handlers.IncreaseQuantity(db))
	router.POST("/cart/:userId/decrement", handlers.DecreaseQuantity(db))
	router.GET("/cart/:userId", handlers.GetCart(db))
	router.POST("/cart/:userId", handlers.AddToCart(db))
	router.DELETE("/cart/:userId/:productId", handlers.RemoveFromCart(db))

	// Роуты для избранного
	router.GET("/favorites/:userId", handlers.GetFavorites(db))                      // Получение избранного пользователя
	router.POST("/favorites/:userId", handlers.AddToFavorites(db))                   // Добавление в избранное
	router.DELETE("/favorites/:userId/:productId", handlers.RemoveFromFavorites(db)) // Удаление из избранного

	//user
	router.POST("/user/register", handlers.RegisterUser(db)) // Регистрация
	router.POST("/user/login", handlers.LoginUser(db))       // Авторизация
	router.GET("/user/:userId", handlers.GetUserInfo(db))    // Информация о пользователе
	router.POST("/user", handlers.CreateUser(db))

	// Роуты для заказов
	router.GET("/orders/:id", handlers.GetOrders(db)) // Получение всех заказов пользователя
	router.POST("/orders", handlers.CreateOrder(db))  // Создание нового заказа
	router.DELETE("/orders/:id", handlers.DeleteOrder(db))

	// Чаты
	router.GET("/chats", handlers.GetChats(db))
	router.GET("/chats/:chatId/messages", handlers.GetMessages(db))
	router.POST("/chats/:chatId/messages", handlers.SendMessage(db))
	router.POST("/chats", handlers.CreateChat(db))
	router.GET("/chats/user/:userId", handlers.GetUserChats(db))

	// Роуты для администраторов
	router.POST("/admin/login", handlers.AdminLogin(db))
	router.GET("/users/admins", handlers.GetAdmins(db)) // Получение списка администраторов

	// Запуск сервера
	router.Run(":8080")
}
