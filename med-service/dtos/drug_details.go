package dtos

type DrugDetails struct {
	ID         string  `json:"id" binding:"required" example:"D007251"`
	Name       string  `json:"name" binding:"required" example:"Tamiflu 30 MG Oral Capsule"`
	Substance  string  `json:"substance" binding:"required" example:"oseltamivir"`
}
