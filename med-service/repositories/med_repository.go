package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"errors"
	"strings"

	commonsErrors "igaku/commons/errors"
	commonsModels "igaku/commons/models"
	medErrors "igaku/med-service/errors"
	"igaku/med-service/models"
)

type MedRepository interface {
	AddPatient(record *commonsModels.PatientRecord) error
	AddDisease(rxNormID string, name string) (*models.Disease, error)
	GetDiseaseByRxNormID(rxNormID string) (*models.Disease, error)
	AddSubstance(rxClassID string, name string, subType string) (*models.Substance, error)
	GetSubstanceByName(name string) (*models.Substance, error)
	AddDrug(rxcui string, name string, substance string) (*models.Drug, error)
	GetDrugByRXCUI(rxcui string) (*models.Drug, error)
	AddMedicalHistoryItem(patientID uuid.UUID, doctorID uuid.UUID) (*models.MedicalHistoryItem, error)
	GetMedicalHistoryItemByPatientID(patientID uuid.UUID) (*models.MedicalHistoryItem, error)
	FindByID(id uuid.UUID) (*commonsModels.PatientRecord, error)
	FindByNationalID(nationalID string) (*commonsModels.PatientRecord, error)
	ValidateUniquePatient(record *commonsModels.PatientRecord) error
	FindBySubstring(name string, offset int, limit int) ([]*models.Disease, error)
	CountBySubstring(name string) (int64, error)
}

type gormMedRepository struct {
	db *gorm.DB
}

func NewGormMedRepository(db *gorm.DB) MedRepository {
	return &gormMedRepository{db: db}
}

func (r *gormMedRepository) AddDisease(
	rxNormID string,
	name string,
) (*models.Disease, error) {
	disease := models.Disease{
		RxNormID: rxNormID,
		Name:     name,
	}

	result := r.db.Where(map[string]interface{}{"rx_norm_id": rxNormID}).FirstOrCreate(&disease)
	if result.Error != nil {
		return nil, &medErrors.DiseaseInsertError{}
	}

	return &disease, nil
}

func (r *gormMedRepository) GetDiseaseByRxNormID(
	rxNormID string,
) (*models.Disease, error) {
	var disease models.Disease

	result := r.db.Where("rx_norm_id = ?", rxNormID).First(&disease)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &medErrors.DiseaseNotFoundError{}
		}
		return nil, &commonsErrors.DatabaseError{}
	}

	return &disease, nil
}

func (r *gormMedRepository) AddMedicalHistoryItem(
	patientID uuid.UUID,
	doctorID uuid.UUID,
) (*models.MedicalHistoryItem, error) {
	item := models.MedicalHistoryItem{
		ID:        uuid.New(),
		PatientID: patientID,
		DoctorID:  doctorID,
	}

	result := r.db.Create(&item)
	if result.Error != nil {
		return nil, &medErrors.MedicalHistoryItemInsertError{}
	}

	return &item, nil
}

func (r *gormMedRepository) GetMedicalHistoryItemByPatientID(
	patientID uuid.UUID,
) (*models.MedicalHistoryItem, error) {
	var item models.MedicalHistoryItem

	result := r.db.Preload("Patient").Preload("Doctor").Where("patient_id = ?", patientID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &medErrors.MedicalHistoryItemNotFoundError{}
		}
		return nil, &commonsErrors.DatabaseError{}
	}

	return &item, nil
}

func (r *gormMedRepository) FindByID(id uuid.UUID) (*commonsModels.PatientRecord, error) {
	var record commonsModels.PatientRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, &medErrors.PatientNotFoundError{}
	}
	return &record, nil
}

func (r *gormMedRepository) AddSubstance(
	rxcui string,
	name string,
	tty string,
) (*models.Substance, error) {
	substance := models.Substance{
		RXCUI: rxcui,
		Name:  name,
		TTY:   tty,
	}

	result := r.db.Where(map[string]interface{}{"rxcui": rxcui}).FirstOrCreate(&substance)
	if result.Error != nil {
		return nil, &medErrors.SubstanceInsertError{}
	}

	return &substance, nil
}

func (r *gormMedRepository) GetSubstanceByName(
	name string,
) (*models.Substance, error) {
	var substance models.Substance

	result := r.db.Where("name = ?", name).First(&substance)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &medErrors.SubstanceNotFoundError{}
		}
		return nil, &commonsErrors.DatabaseError{}
	}

	return &substance, nil
}

func (r *gormMedRepository) AddDrug(
	rxcui string,
	name string,
	substance string,
) (*models.Drug, error) {
	drug := models.Drug{
		RXCUI:     rxcui,
		Name:      name,
		Substance: substance,
	}

	result := r.db.Where(map[string]interface{}{"rxcui": rxcui}).FirstOrCreate(&drug)
	if result.Error != nil {
		return nil, &medErrors.DrugInsertError{}
	}

	return &drug, nil
}

func (r *gormMedRepository) GetDrugByRXCUI(rxcui string) (*models.Drug, error) {
	var drug models.Drug

	result := r.db.Where("rxcui = ?", rxcui).First(&drug)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, &medErrors.DrugNotFoundError{}
		}
		return nil, &commonsErrors.DatabaseError{}
	}

	return &drug, nil
}

func (r *gormMedRepository) FindByNationalID(
	nationalID string,
) (*commonsModels.PatientRecord, error) {
	var record commonsModels.PatientRecord
	err := r.db.Where(&commonsModels.PatientRecord{NationalID: nationalID}).First(&record).Error
	if err != nil {
		return nil, &medErrors.PatientNotFoundError{}
	}
	return &record, nil
}

func (r *gormMedRepository) ValidateUniquePatient(record *commonsModels.PatientRecord) error {
	var existingPatient commonsModels.PatientRecord
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

func (r *gormMedRepository) AddPatient(record *commonsModels.PatientRecord) error {
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
		return nil, &medErrors.OffsetNegativeError{}
	}
	if limit < 0 {
		return nil, &medErrors.LimitNegativeError{}
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
		return nil, &medErrors.DiseaseNotFoundError{}
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
		return 0, &medErrors.DiseaseNotFoundError{}
	}
	return count, nil
}
