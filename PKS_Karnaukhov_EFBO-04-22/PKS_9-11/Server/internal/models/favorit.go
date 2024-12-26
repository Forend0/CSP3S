package models

import "time"

// Cart представляет структуру корзины
type Favorite struct {
	FavoriteID int       `db:"favorite_id" json:"favorite_id"`
	UserID     int       `db:"user_id" json:"user_id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
