package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"strings"

	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
	"igaku/med-service/errors"
)

type PatientRepository interface {
	AddPatient(record *models.PatientRecord) error
	FindByID(id uuid.UUID) (*models.PatientRecord, error)
	ValidateUniquePatient(record *models.PatientRecord) error
}

type gormPatientRepository struct {
	db *gorm.DB
}

func NewGormPatientRepository(db *gorm.DB) PatientRepository {
	return &gormPatientRepository{db: db}
}

func (r *gormPatientRepository) FindByID(id uuid.UUID) (*models.PatientRecord, error) {
	var record models.PatientRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, &errors.PatientNotFoundError{}
	}
	return &record, nil
}

func (r *gormPatientRepository) ValidateUniquePatient(record *models.PatientRecord) error {
	var existingPatient models.PatientRecord
	result := r.db.
		Where("id = ?", &record.ID).
		Or("national_id = ?", &record.NationalID).
		First(&existingPatient)

	if result.Error == nil {
		if existingPatient.ID == record.ID {
			return &commonsErrors.DuplicatedIDError{
				ID: record.ID,
			}
		}
		if existingPatient.NationalID == record.NationalID {
			return &commonsErrors.DuplicatedNationalIDError{
				NationalID: record.NationalID,
			}
		}
	}

	return nil
}

func (r *gormPatientRepository) AddPatient(record *models.PatientRecord) error {
    err := r.db.Create(record).Error
    if err == nil {
        return nil
    }

    errMsg := err.Error()

    if strings.Contains(errMsg, "duplicate key") {
        switch {
        case strings.Contains(errMsg, "patient_records_pkey"):
            return &commonsErrors.DuplicatedIDError{
                ID: record.ID,
            }
        case strings.Contains(errMsg, "idx_patient_records_national_id"):
            return &commonsErrors.DuplicatedNationalIDError{
                NationalID: record.NationalID,
            }
        }
    }

    if strings.Contains(errMsg, "chk_patient_records_national_id") {
        return &commonsErrors.InvalidNationalIDError{
            NationalID: record.NationalID,
        }
    }

    return &commonsErrors.DatabaseError{}
}
