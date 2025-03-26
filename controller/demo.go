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
	router.POST("auth/change-password", changePassword)
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

// เพิ่มฟังก์ชันสำหรับเปลี่ยนรหัสผ่าน
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
