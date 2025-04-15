package dto

type LoginCredentials struct {
	Username string `json:"username" binding:"required" example:"jdoe"`
	Password string `json:"password" binding:"required" example:"P@ssw0rd!"`
}
