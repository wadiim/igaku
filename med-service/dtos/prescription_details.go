package dtos

type PrescriptionDetails struct {
	Patient  PatientDetails  `json:"patient" binding:"required"`
	Disease  DiseaseDetails  `json:"disease" binding:"required"`
	Drugs    []DrugDetails   `json:"drugs" binding:"required"`
}
