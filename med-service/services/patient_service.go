package services

import (
	"igaku/commons/models"
	"igaku/med-service/repositories"
)

type PatientService interface {
	ValidateUniquePatient(record *models.PatientRecord) error
	CreatePatient(data *models.PatientRecord) error
}

type patientService struct {
	repo repositories.PatientRepository
}

func NewPatientService(repo repositories.PatientRepository) PatientService {
	return &patientService{repo}
}

func (s *patientService) ValidateUniquePatient(record *models.PatientRecord) error {
	err := s.repo.ValidateUniquePatient(record)

	return err
}

func (s *patientService) CreatePatient(data *models.PatientRecord) error {
	err := s.repo.AddPatient(data)

	return err
}
