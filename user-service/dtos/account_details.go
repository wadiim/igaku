package dtos

type AccountDetails struct {
	Username	string `json:"username" binding:"required" example:"jdoe"`
	Email		string `json:"email" binding:"required" example:"jdoe@mail.com"`
	Role		string `json:"role" binding:"required" example:"patient"`
}
