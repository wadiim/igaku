package repositories

import (
	"gorm.io/gorm"

	"strings"

	models "igaku/commons/models"
	errors "igaku/commons/errors"
	igakuErrors "igaku/med-service/errors"
)

type DiseaseRepository interface {
	FindBySubstring(name string, offset int, limit int) ([]*models.Disease, error)
	CountBySubstring(name string) (int64, error)
}

type gormDiseaseRepository struct {
	db *gorm.DB
}

func NewGormDiseaseRepository(db *gorm.DB) DiseaseRepository {
	return &gormDiseaseRepository{db: db}
}

func (r *gormDiseaseRepository) FindBySubstring(
	name string,
	offset int,
	limit int,
) ([]*models.Disease, error) {
	if offset < 0 {
		return nil, &igakuErrors.OffsetNegativeError{}
	}
	if limit < 0 {
		return nil, &igakuErrors.LimitNegativeError{}
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
		return nil, &errors.DatabaseError{}
	}
	if len(diseases) == 0 {
		return nil, &igakuErrors.DiseaseNotFoundError{}
	}
	return diseases, nil
}

func (r *gormDiseaseRepository) CountBySubstring(name string) (int64, error) {
	var count int64
	// Converting `name` to lowercase is necessary to make sure that names provided
	// in uppercase are also included in the results
	result := r.db.
		Model(&models.Disease{}).
		Where("LOWER(name) LIKE ?", "%"+strings.ToLower(name)+"%").
		Count(&count)
	if result.Error != nil {
		return 0, &errors.DatabaseError{}
	}
	if count == 0 {
		return 0, &igakuErrors.DiseaseNotFoundError{}
	}
	return count, nil
}
