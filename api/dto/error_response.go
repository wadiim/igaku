package dto

type ErrorResponse struct {
	Message string `json:"error" example:"Specific error message"`
}
