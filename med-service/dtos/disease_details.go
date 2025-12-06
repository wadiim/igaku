package dtos

type DiseaseDetails struct {
	ID	string `json:"id" binding:"required" example:"6288b3bd-959f-4b57-a26e-11688e26ce5c"`
	RxNormID	string `json:"rx_norm_id" binding:"required" example:"D000111"`
	Name	string `json:"name" binding:"required" example:"Pneumonia"`
}
