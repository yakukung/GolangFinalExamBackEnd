package models

import "time"

// CartResponse โครงสร้างสำหรับการแสดงผลรถเข็นทั้งหมดของลูกค้า
type CartResponse struct {
	CustomerID int        `json:"customer_id"`
	Carts      []CartInfo `json:"carts"`
}

// CartInfo โครงสร้างข้อมูลรถเข็น
type CartInfo struct {
	CartID    int        `json:"cart_id"`
	CartName  string     `json:"cart_name"`
	CreatedAt time.Time  `json:"created_at"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
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

// Cart โครงสร้างข้อมูลรถเข็นสำหรับการจัดการในฐานข้อมูล
type Cart struct {
	CartID     int       `json:"cart_id"`
	CustomerID int       `json:"customer_id"`
	CartName   string    `json:"cart_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CartItemDB โครงสร้างข้อมูลสินค้าในรถเข็นสำหรับการจัดการในฐานข้อมูล
type CartItemDB struct {
	CartItemID int       `json:"cart_item_id"`
	CartID     int       `json:"cart_id"`
	ProductID  int       `json:"product_id"`
	Quantity   int       `json:"quantity"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Product โครงสร้างข้อมูลสินค้า
type Product struct {
	ProductID     int       `json:"product_id"`
	ProductName   string    `json:"product_name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SearchAndAddToCartRequest โครงสร้างข้อมูลสำหรับการค้นหาและเพิ่มสินค้าลงรถเข็น
type SearchAndAddToCartRequest struct {
	SearchDescription string  `json:"search_description"`
	MinPrice          float64 `json:"min_price"`
	MaxPrice          float64 `json:"max_price"`
	ProductID         int     `json:"product_id"`
	Quantity          int     `json:"quantity"`
	CartName          string  `json:"cart_name"`
}
