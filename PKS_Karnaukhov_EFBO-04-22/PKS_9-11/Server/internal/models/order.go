package models

import "time"

// Order представляет заказ
type Order struct {
	OrderID   int        `json:"order_id" db:"order_id"`     // ID заказа
	UserID    int        `json:"user_id" db:"user_id"`       // ID пользователя
	Total     float64    `json:"total" db:"total"`           // Общая сумма заказа
	Status    string     `json:"status" db:"status"`         // Статус заказа
	CreatedAt time.Time  `json:"created_at" db:"created_at"` // Дата создания заказа
	Products  []struct { // Список продуктов в заказе
		ProductID int `json:"product_id"` // ID продукта
		Quantity  int `json:"quantity"`   // Количество
	} `json:"products"`
	DeliveryAddress string `json:"delivery_address" db:"delivery_address"` // Адрес доставки
}
