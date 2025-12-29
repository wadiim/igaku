package services

import (
	"log"
	"igaku/commons/models"
	"igaku/med-service/clients"
	"igaku/med-service/dtos"
	"igaku/med-service/repositories"
)

type PatientService interface {
	ValidateUniquePatient(record *models.PatientRecord) error
	CreatePatient(data *models.PatientRecord) error
	GetPatientByNationalID(nationalID string) (*dtos.PatientDetails, error)
}

type patientService struct {
	userClient clients.UserClient
	repo repositories.PatientRepository
}

func NewPatientService(
	userClient clients.UserClient,
	repo repositories.PatientRepository,
) PatientService {
	return &patientService{userClient, repo}
}

func (s *patientService) GetPatientByNationalID(nationalID string) (*dtos.PatientDetails, error) {
	record, err := s.repo.FindByNationalID(nationalID)
	if err != nil {
		return nil, err
	}

	log.Printf("Record: %v", record)
	patient, err := s.userClient.FindByID(record.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("Patient: %v", patient)

	patientDetails := &dtos.PatientDetails{
		Username: patient.Username,
		Email: patient.Email,
		NationalID: record.NationalID,
	}

	return patientDetails, nil
}

func (s *patientService) ValidateUniquePatient(record *models.PatientRecord) error {
	err := s.repo.ValidateUniquePatient(record)

	return err
}

func (s *patientService) CreatePatient(data *models.PatientRecord) error {
	err := s.repo.AddPatient(data)

	return err
}
