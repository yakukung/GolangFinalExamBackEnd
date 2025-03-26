package models

import "time"

// CartResponse โครงสร้างสำหรับการแสดงผลรถเข็นทั้งหมดของลูกค้า
type CartResponse struct {
	CustomerID int        `json:"customer_id"`
	Carts      []CartInfo `json:"carts"`
}

// CartInfo โครงสร้างข้อมูลรถเข็น
type CartInfo struct {
	CartID    int         `json:"cart_id"`
	CartName  string      `json:"cart_name"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []CartItem  `json:"items"`
	Total     float64     `json:"total"`
}

// CartItem โครงสร้างข้อมูลสินค้าในรถเข็น
type CartItem struct {
	CartItemID  int     `json:"cart_item_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}
