package dtos

type RegistrationFields struct {
	Username string `json:"username" binding:"required" example:"jdoe"`
	Password string `json:"password" binding:"required" example:"P@ssw0rd!"`
}
