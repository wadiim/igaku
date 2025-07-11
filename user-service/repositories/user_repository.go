package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"fmt"
	"strings"

	"igaku/user-service/errors"
	"igaku/user-service/utils"
	commonsErrors "igaku/commons/errors"
	"igaku/commons/models"
)

type UserRepository interface {
	FindByID(id uuid.UUID) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindAll(
		offset, limit int,
		orderBy models.UserOrderableField, orderMethod utils.Ordering,
	) ([]models.User, error)
	CountAll() (int64, error)
	Persist(user *models.User) (error)
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
		return nil, &commonsErrors.UserNotFoundError{}
	}
	return &usr, nil
}

func (r *gormUserRepository) FindByUsername(username string) (*models.User, error) {
	var usr models.User
	err := r.db.Where("username = ?", username).First(&usr).Error
	if err != nil {
		return nil, &commonsErrors.UserNotFoundError{}
	}
	return &usr, nil
}

func (r *gormUserRepository) FindAll(
	offset, limit int,
	orderBy models.UserOrderableField, orderMethod utils.Ordering,
) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Order(fmt.Sprintf("%s %s", string(orderBy), string(orderMethod))).
		Offset(offset).
		Limit(limit).
		Find(&users).
		Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *gormUserRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *gormUserRepository) Persist(user *models.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "users_pkey") {
				// TODO: Use `DuplicatedIDError`
				return &errors.InvalidUserError{
					"Duplicated ID",
				}
			} else if strings.Contains(err.Error(), "users_username") {
				// TODO: Use `DuplicatedUsernameError`
				return &errors.InvalidUserError{
					"Duplicated username",
				}
			} else if strings.Contains(err.Error(), "idx_users_email") {
				return &commonsErrors.DuplicatedEmailError{}
			}
		}
		return err
	}
	return err
}
