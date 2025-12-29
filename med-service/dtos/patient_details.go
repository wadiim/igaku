package dtos

type PatientDetails struct {
	Username	string `json:"username" binding:"required" example:"jdoe"`
	Email	string `json:"email" binding:"required" example:"jdoe@mail.com"`
	NationalID	string `json:"national_id" binding:"required" example:"44051401458"`
}
