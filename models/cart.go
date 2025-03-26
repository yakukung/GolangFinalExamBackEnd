package models

import (
	"time"
)

type Cart struct {
	CartID     int        `json:"cart_id" db:"cart_id"`
	CustomerID int        `json:"customer_id" db:"customer_id"`
	CartName   string     `json:"cart_name" db:"cart_name"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	Items      []CartItem `json:"items,omitempty" db:"-"` // For joining with cart items
}
