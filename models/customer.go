package models

import (
	"time"
)

type Customer struct {
	CustomerID  int       `json:"customer_id" db:"customer_id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	Email       string    `json:"email" db:"email"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Address     string    `json:"address" db:"address"`
	Password    string    `json:"-" db:"password"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
