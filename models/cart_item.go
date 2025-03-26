package models

import (
	"time"
)

type CartItem struct {
	CartItemID int       `json:"cart_item_id" db:"cart_item_id"`
	CartID     int       `json:"cart_id" db:"cart_id"`
	ProductID  int       `json:"product_id" db:"product_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Product    *Product  `json:"product,omitempty" db:"-"` // For joining with product details
}
