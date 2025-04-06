package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"igaku/models"
)

type OrganizationRepository interface {
	FindByID(id uuid.UUID) (*models.Organization, error)
}

type gormOrganizationRepository struct {
	db *gorm.DB
}

func NewGormOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &gormOrganizationRepository{db: db}
}

func (r *gormOrganizationRepository) FindByID(id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}
