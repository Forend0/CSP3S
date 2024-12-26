package models

type Product struct {
	ProductID   int     `db:"product_id" json:"product_id"`
	Name        string  `db:"name" json:"name"`
	Description string  `db:"description" json:"description"`
	Price       float64 `db:"price" json:"price"`
	Stock       int     `db:"stock" json:"stock"`
	ImageURL    string  `db:"image_url" json:"image_url"`
}
