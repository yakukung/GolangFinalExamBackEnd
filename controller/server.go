package controller

import (
	"fmt"
	"go-final/database"
	"go-final/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func StartServer() {
	database.ConnectDatabase()
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API is now working",
		})
	})

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, world!",
		})
	})

	// Register other controllers
	DemoController(router)

	// Start the server
	router.Run(":8080")
}

// ฟังก์ชันสำหรับเข้ารหัสรหัสผ่านของผู้ใช้ที่ระบุ
func EncryptPassword(email string) (string, error) {
	var customer models.Customer

	// ค้นหาผู้ใช้จากอีเมล
	if err := database.DB.Table("customer").Where("email = ?", email).First(&customer).Error; err != nil {
		return "", fmt.Errorf("error fetching customer with email: %s, %v", email, err)
	}
	// ตรวจสอบว่ารหัสผ่านเป็น bcrypt hash หรือไม่
	if len(customer.Password) > 0 && customer.Password[0] != '$' {
		// เก็บรหัสผ่านเดิม
		originalPassword := customer.Password

		// เข้ารหัสรหัสผ่านด้วย bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(customer.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("error hashing password for customer ID: %d, %v", customer.CustomerID, err)
		}

		// อัปเดตรหัสผ่านในฐานข้อมูล
		if err := database.DB.Table("customer").Where("customer_id = ?", customer.CustomerID).Update("password", string(hashedPassword)).Error; err != nil {
			return "", fmt.Errorf("error updating password for customer ID: %d, %v", customer.CustomerID, err)
		}

		fmt.Println("Password encrypted for customer ID:", customer.CustomerID)
		return originalPassword, nil
	}

	return customer.Password, nil
}
