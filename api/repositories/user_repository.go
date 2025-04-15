package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"igaku/models"
)

type UserRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var usr models.User
	err := r.db.First(&usr, id).Error
	if err != nil {
		return nil, err
	}
	return &usr, nil
}

func (r *gormUserRepository) FindByUsername(username string) (*models.User, error) {
	var usr models.User
	err := r.db.Where("username = ?", username).First(&usr).Error
	if err != nil {
		return nil, err
	}
	return &usr, nil
}
