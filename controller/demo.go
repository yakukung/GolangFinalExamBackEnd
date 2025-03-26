package controller

import (
	"fmt"
	"go-final/database"
	"go-final/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func DemoController(router *gin.Engine) {
	router.POST("auth/login", login)
	router.POST("auth/change-password", changePassword)
	router.GET("customer/:customer_id/carts", getCustomerCarts)
	router.GET("products/search", searchProducts)
	router.POST("customer/:customer_id/add-to-cart", addToCart)
	// router.POST("customer/:customer_id/add-to-cart", searchAndAddToCart)
}

// ล็อคอิน
func login(c *gin.Context) {
	var request models.LoginRequest
	var customer models.Customer
	var originalPassword string

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	if request.Password == "" || request.Email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// เข้ารหัสรหัสผ่านของผู้ใช้ที่กำลังล็อกอิน และเก็บรหัสผ่านเดิม
	var err error
	originalPassword, err = EncryptPassword(request.Email)
	if err != nil {
		fmt.Println("Error encrypting password:", err)
	}

	// Log รหัสผ่านเดิม
	fmt.Println("Original password:", originalPassword)
	// ค้นหาผ่านจากอีเมล (ดึงข้อมูลใหม่หลังจากเข้ารหัสแล้ว)
	if err := database.DB.Table("customer").Where("email = ?", request.Email).First(&customer).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Log รหัสผ่านที่เข้ารหัสแล้ว
	fmt.Println("Encrypted password:", customer.Password)
	// ตรวจสอบรหัสผ่านด้วย bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// ล็อกอินสำเร็จ ให้กลับไปใช้รหัสผ่านเดิม
	if originalPassword != "" && originalPassword[0] != '$' {
		if err := database.DB.Table("customer").Where("customer_id = ?", customer.CustomerID).
			Update("password", originalPassword).Error; err != nil {
			fmt.Println("Error reverting to original password for customer ID:", customer.CustomerID, err)
		} else {
			fmt.Println("Reverted to original password for customer ID:", customer.CustomerID)
		}
	}

	// ไม่ส่งรหัสผ่านกลับไป
	customer.Password = ""
	c.JSON(http.StatusOK, customer)

}

// เปลี่ยนรหัสผ่าน
func changePassword(c *gin.Context) {
	var request models.ChangePasswordRequest
	var customer models.Customer

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// ตรวจสอบว่ามีข้อมูลครบถ้วนหรือไม่
	if request.CustomerID == 0 || request.OldPassword == "" || request.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID, old password and new password are required"})
		return
	}

	// ค้นหาผู้ใช้จาก customer_id
	if err := database.DB.Table("customer").Where("customer_id = ?", request.CustomerID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// ดึงรหัสผ่านปัจจุบันจากฐานข้อมูล
	var currentPassword string
	if err := database.DB.Table("customer").Select("password").Where("customer_id = ?", request.CustomerID).Row().Scan(&currentPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current password"})
		return
	}

	// ตรวจสอบรหัสผ่านเก่า - ตรวจสอบทั้งกรณีที่เป็น plain text และ bcrypt hash
	var passwordMatch bool

	// กรณีที่รหัสผ่านเป็น bcrypt hash (ขึ้นต้นด้วย $)
	if len(currentPassword) > 0 && currentPassword[0] == '$' {
		err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(request.OldPassword))
		passwordMatch = (err == nil)
	} else {
		// กรณีที่รหัสผ่านเป็น plain text
		passwordMatch = (currentPassword == request.OldPassword)
	}

	if !passwordMatch {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// เข้ารหัสรหัสผ่านใหม่
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt new password"})
		return
	}

	// อัปเดตรหัสผ่านในฐานข้อมูล
	if err := database.DB.Table("customer").Where("customer_id = ?", customer.CustomerID).
		Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// แสดงรายการตะกร้าผู้ใช้
func getCustomerCarts(c *gin.Context) {
	customerID := c.Param("customer_id")

	// ตรวจสอบว่ามี customer_id หรือไม่
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID is required"})
		return
	}

	// ตรวจสอบว่ามีลูกค้าในระบบหรือไม่
	var customer models.Customer
	if err := database.DB.Table("customer").Where("customer_id = ?", customerID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// ดึงข้อมูลรถเข็นทั้งหมดของลูกค้า
	var carts []struct {
		CartID    int       `json:"cart_id"`
		CartName  string    `json:"cart_name"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := database.DB.Table("cart").
		Select("cart_id, cart_name, created_at").
		Where("customer_id = ?", customerID).
		Find(&carts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve carts"})
		return
	}

	// สร้างข้อมูลการตอบกลับ
	response := models.CartResponse{
		CustomerID: customer.CustomerID,
		Carts:      []models.CartInfo{},
	}

	// ดึงข้อมูลสินค้าในแต่ละรถเข็น
	for _, cart := range carts {
		cartInfo := models.CartInfo{
			CartID:    cart.CartID,
			CartName:  cart.CartName,
			CreatedAt: cart.CreatedAt,
			Items:     []models.CartItem{},
			Total:     0,
		}

		// ดึงข้อมูลสินค้าในรถเข็น
		var items []struct {
			CartItemID  int     `json:"cart_item_id"`
			ProductID   int     `json:"product_id"`
			ProductName string  `json:"product_name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Quantity    int     `json:"quantity"`
		}

		if err := database.DB.Table("cart_item").
			Select("cart_item.cart_item_id, product.product_id, product.product_name, product.description, product.price, cart_item.quantity").
			Joins("JOIN product ON cart_item.product_id = product.product_id").
			Where("cart_item.cart_id = ?", cart.CartID).
			Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart items"})
			return
		}

		// คำนวณราคารวมของแต่ละรายการและรถเข็น
		for _, item := range items {
			subtotal := float64(item.Quantity) * item.Price
			cartItem := models.CartItem{
				CartItemID:  item.CartItemID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Description: item.Description,
				Price:       item.Price,
				Quantity:    item.Quantity,
				Subtotal:    subtotal,
			}
			cartInfo.Items = append(cartInfo.Items, cartItem)
			cartInfo.Total += subtotal
		}

		response.Carts = append(response.Carts, cartInfo)
	}

	c.JSON(http.StatusOK, response)
}

// // ค้นหาสินค้าและเพิ่มสินค้าลงรถเข็น
// func searchAndAddToCart(c *gin.Context) {
// 	customerID := c.Param("customer_id")

// 	// ตรวจสอบว่ามี customer_id หรือไม่
// 	if customerID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID is required"})
// 		return
// 	}

// 	// ตรวจสอบว่ามีลูกค้าในระบบหรือไม่
// 	var customer models.Customer
// 	if err := database.DB.Table("customer").Where("customer_id = ?", customerID).First(&customer).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
// 		return
// 	}

// 	// รับข้อมูลการค้นหาและการเพิ่มสินค้าลงรถเข็น
// 	var request struct {
// 		SearchDescription string  `json:"search_description"`
// 		MinPrice          float64 `json:"min_price"`
// 		MaxPrice          float64 `json:"max_price"`
// 		ProductID         int     `json:"product_id"`
// 		Quantity          int     `json:"quantity"`
// 		CartName          string  `json:"cart_name"`
// 	}

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
// 		return
// 	}

// 	// ตรวจสอบข้อมูลที่จำเป็น
// 	if request.ProductID == 0 || request.Quantity <= 0 || request.CartName == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID, quantity, and cart name are required"})
// 		return
// 	}

// 	// ค้นหาสินค้าตามเงื่อนไข
// 	var products []models.Product
// 	query := database.DB.Table("product").Where("product_id = ?", request.ProductID)

// 	// เพิ่มเงื่อนไขการค้นหาตามคำอธิบายสินค้า (ถ้ามี)
// 	if request.SearchDescription != "" {
// 		query = query.Where("description LIKE ?", "%"+request.SearchDescription+"%")
// 	}

// 	// เพิ่มเงื่อนไขการค้นหาตามช่วงราคา (ถ้ามี)
// 	if request.MinPrice > 0 {
// 		query = query.Where("price >= ?", request.MinPrice)
// 	}
// 	if request.MaxPrice > 0 {
// 		query = query.Where("price <= ?", request.MaxPrice)
// 	}

// 	// ดึงข้อมูลสินค้า
// 	if err := query.Find(&products).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
// 		return
// 	}

// 	// ตรวจสอบว่าพบสินค้าหรือไม่
// 	if len(products) == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "No products found matching the criteria"})
// 		return
// 	}

// 	// ตรวจสอบว่ามีรถเข็นตามชื่อที่กำหนดหรือไม่
// 	var cart models.Cart
// 	err := database.DB.Table("cart").Where("customer_id = ? AND cart_name = ?", customerID, request.CartName).First(&cart).Error

// 	// ถ้าไม่มีรถเข็น ให้สร้างใหม่
// 	if err != nil {
// 		cart = models.Cart{
// 			CustomerID: customer.CustomerID,
// 			CartName:   request.CartName,
// 		}

// 		if err := database.DB.Table("cart").Create(&cart).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new cart"})
// 			return
// 		}
// 	}

// 	// ตรวจสอบว่ามีสินค้านี้ในรถเข็นอยู่แล้วหรือไม่
// 	var existingItem models.CartItemDB
// 	err = database.DB.Table("cart_item").
// 		Where("cart_id = ? AND product_id = ?", cart.CartID, request.ProductID).
// 		First(&existingItem).Error

// 	// ถ้ามีสินค้านี้อยู่แล้ว ให้เพิ่มจำนวน
// 	if err == nil {
// 		// อัปเดตจำนวนสินค้า
// 		newQuantity := existingItem.Quantity + request.Quantity

// 		if err := database.DB.Table("cart_item").
// 			Where("cart_item_id = ?", existingItem.CartItemID).
// 			Update("quantity", newQuantity).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"message":      "Product quantity updated in cart",
// 			"cart_id":      cart.CartID,
// 			"cart_name":    cart.CartName,
// 			"product_id":   request.ProductID,
// 			"new_quantity": newQuantity,
// 		})
// 		return
// 	}

// 	// ถ้ายังไม่มีสินค้านี้ในรถเข็น ให้เพิ่มใหม่
// 	newItem := models.CartItemDB{
// 		CartID:    cart.CartID,
// 		ProductID: request.ProductID,
// 		Quantity:  request.Quantity,
// 	}

// 	if err := database.DB.Table("cart_item").Create(&newItem).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":    "Product added to cart successfully",
// 		"cart_id":    cart.CartID,
// 		"cart_name":  cart.CartName,
// 		"product_id": request.ProductID,
// 		"quantity":   request.Quantity,
// 	})
// }

// ค้นหาสินค้า
func searchProducts(c *gin.Context) {
	// รับพารามิเตอร์การค้นหาจาก query string
	searchDescription := c.Query("description")
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")

	// แปลงค่าราคาจาก string เป็น float64
	var minPrice, maxPrice float64
	var err error

	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price format"})
			return
		}
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price format"})
			return
		}
	}

	// สร้าง query สำหรับค้นหาสินค้า
	query := database.DB.Table("product")

	// เพิ่มเงื่อนไขการค้นหาตามคำอธิบายสินค้า (ถ้ามี)
	if searchDescription != "" {
		query = query.Where("description LIKE ? OR product_name LIKE ?",
			"%"+searchDescription+"%", "%"+searchDescription+"%")
	}

	// เพิ่มเงื่อนไขการค้นหาตามช่วงราคา (ถ้ามี)
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}

	// ดึงข้อมูลสินค้า
	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products"})
		return
	}

	// ตรวจสอบว่าพบสินค้าหรือไม่
	if len(products) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No products found matching the criteria"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"count":    len(products),
	})
}

// เพิ่มสินค้าลงรถเข็น
func addToCart(c *gin.Context) {
	customerID := c.Param("customer_id")

	// ตรวจสอบว่ามี customer_id หรือไม่
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID is required"})
		return
	}

	// ตรวจสอบว่ามีลูกค้าในระบบหรือไม่
	var customer models.Customer
	if err := database.DB.Table("customer").Where("customer_id = ?", customerID).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// รับข้อมูลการเพิ่มสินค้าลงรถเข็น
	var request struct {
		ProductID int    `json:"product_id"`
		Quantity  int    `json:"quantity"`
		CartName  string `json:"cart_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if request.ProductID == 0 || request.Quantity <= 0 || request.CartName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID, quantity, and cart name are required"})
		return
	}

	// ตรวจสอบว่ามีสินค้านี้ในระบบหรือไม่
	var product models.Product
	if err := database.DB.Table("product").Where("product_id = ?", request.ProductID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// ตรวจสอบว่ามีรถเข็นตามชื่อที่กำหนดหรือไม่
	var cart models.Cart
	err := database.DB.Table("cart").Where("customer_id = ? AND cart_name = ?", customerID, request.CartName).First(&cart).Error

	// ถ้าไม่มีรถเข็น ให้สร้างใหม่
	if err != nil {
		cart = models.Cart{
			CustomerID: customer.CustomerID,
			CartName:   request.CartName,
		}

		if err := database.DB.Table("cart").Create(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new cart"})
			return
		}
	}

	// ตรวจสอบว่ามีสินค้านี้ในรถเข็นอยู่แล้วหรือไม่
	var existingItem models.CartItemDB
	err = database.DB.Table("cart_item").
		Where("cart_id = ? AND product_id = ?", cart.CartID, request.ProductID).
		First(&existingItem).Error

	// ถ้ามีสินค้านี้อยู่แล้ว ให้เพิ่มจำนวน
	if err == nil {
		// อัปเดตจำนวนสินค้า
		newQuantity := existingItem.Quantity + request.Quantity

		if err := database.DB.Table("cart_item").
			Where("cart_item_id = ?", existingItem.CartItemID).
			Update("quantity", newQuantity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item quantity"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Product quantity updated in cart",
			"cart_id":      cart.CartID,
			"cart_name":    cart.CartName,
			"product_id":   request.ProductID,
			"product_name": product.ProductName,
			"price":        product.Price,
			"new_quantity": newQuantity,
		})
		return
	}

	// ถ้ายังไม่มีสินค้านี้ในรถเข็น ให้เพิ่มใหม่
	newItem := models.CartItemDB{
		CartID:    cart.CartID,
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}

	if err := database.DB.Table("cart_item").Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Product added to cart successfully",
		"cart_id":      cart.CartID,
		"cart_name":    cart.CartName,
		"product_id":   request.ProductID,
		"product_name": product.ProductName,
		"price":        product.Price,
		"quantity":     request.Quantity,
	})
}
