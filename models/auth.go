package models

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	CustomerID int    `json:"customer_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Token      string `json:"token,omitempty"`
}

// ChangePasswordRequest โครงสร้างสำหรับคำขอเปลี่ยนรหัสผ่าน
type ChangePasswordRequest struct {
	CustomerID  int    `json:"customer_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
