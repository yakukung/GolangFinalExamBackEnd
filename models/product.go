package models

import (
	"gorm.io/gorm"
	"time"
)

type Product struct {
	gorm.Model
	ProductID     int       `json:"product_id" gorm:"primaryKey"`
	ProductName   string    `json:"product_name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
