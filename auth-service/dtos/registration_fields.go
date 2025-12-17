package dtos

type RegistrationFields struct {
	Username string `json:"username" binding:"required" example:"jdoe"`
	Email string `json:"email" binding:"required" example:"jdoe@mail.com"`
	NationalID string `json:"national_id" binding:"required" example:"51011664198"`
	Password string `json:"password" binding:"required" example:"P@ssw0rd!"`
}
