package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"strings"

	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
	"igaku/med-service/errors"
)

type MedRepository interface {
	AddPatient(record *models.PatientRecord) error
	FindByID(id uuid.UUID) (*models.PatientRecord, error)
	FindByNationalID(nationalID string) (*models.PatientRecord, error)
	ValidateUniquePatient(record *models.PatientRecord) error
	FindBySubstring(name string, offset int, limit int) ([]*models.Disease, error)
	CountBySubstring(name string) (int64, error)
}

type gormMedRepository struct {
	db *gorm.DB
}

func NewGormMedRepository(db *gorm.DB) MedRepository {
	return &gormMedRepository{db: db}
}

func (r *gormMedRepository) FindByID(id uuid.UUID) (*models.PatientRecord, error) {
	var record models.PatientRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, &errors.PatientNotFoundError{}
	}
	return &record, nil
}

func (r *gormMedRepository) FindByNationalID(
	nationalID string,
) (*models.PatientRecord, error) {
	var record models.PatientRecord
	err := r.db.Where(&models.PatientRecord{NationalID: nationalID}).First(&record).Error
	if err != nil {
		return nil, &errors.PatientNotFoundError{}
	}
	return &record, nil
}

func (r *gormMedRepository) ValidateUniquePatient(record *models.PatientRecord) error {
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

func (r *gormMedRepository) AddPatient(record *models.PatientRecord) error {
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

func (r *gormMedRepository) FindBySubstring(
	name string,
	offset int,
	limit int,
) ([]*models.Disease, error) {
	if offset < 0 {
		return nil, &errors.OffsetNegativeError{}
	}
	if limit < 0 {
		return nil, &errors.LimitNegativeError{}
	}
	var diseases []*models.Disease
	// Converting `name` to lowercase is necessary to make sure that names provided
	// in uppercase are also included in the results
	result := r.db.
		Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%").
		Limit(limit).
		Offset(offset).
		Find(&diseases)
	if result.Error != nil {
		return nil, &commonsErrors.DatabaseError{}
	}
	if len(diseases) == 0 {
		return nil, &errors.DiseaseNotFoundError{}
	}
	return diseases, nil
}

func (r *gormMedRepository) CountBySubstring(name string) (int64, error) {
	var count int64
	// Converting `name` to lowercase is necessary to make sure that names provided
	// in uppercase are also included in the results
	result := r.db.
		Model(&models.Disease{}).
		Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%").
		Count(&count)
	if result.Error != nil {
		return 0, &commonsErrors.DatabaseError{}
	}
	if count == 0 {
		return 0, &errors.DiseaseNotFoundError{}
	}
	return count, nil
}
