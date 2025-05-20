package dtos

type RegistrationFields struct {
	Username string `json:"username" binding:"required" example:"jdoe"`
	Email string `json:"email" binding:"required" example:"jdoe@mail.com"`
	Password string `json:"password" binding:"required" example:"P@ssw0rd!"`
}
