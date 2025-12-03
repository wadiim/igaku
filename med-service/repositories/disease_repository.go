package repositories

import (
	"gorm.io/gorm"

	models "igaku/commons/models"
	errors "igaku/commons/errors"
	igakuErrors "igaku/med-service/errors"
)

type DiseaseRepository interface {
	FindByName(name string) ([]*models.Disease, error)
}

type gormDiseaseRepository struct {
	db *gorm.DB
}

func NewGormDiseaseRepository(db *gorm.DB) DiseaseRepository {
	return &gormDiseaseRepository{db: db}
}

func (r *gormDiseaseRepository) FindByName(name string) ([]*models.Disease, error) {
	var diseases []*models.Disease
	result := r.db.Where("name LIKE ?", "%"+name+"%").Find(&diseases)
	if result.Error != nil {
		return nil, &errors.DatabaseError{}
	}
	if len(diseases) == 0 {
		return nil, &igakuErrors.DiseaseNotFoundError{}
	}
	return diseases, nil
}

