package controller

import (
	"fmt"
	"go-final/database"
	"go-final/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func DemoController(router *gin.Engine) {
	router.POST("auth/login", login)
}

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
	// ค้นหาผู้ใช้จากอีเมล (ดึงข้อมูลใหม่หลังจากเข้ารหัสแล้ว)
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
