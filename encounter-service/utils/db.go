package utils

import (
	"gorm.io/gorm"

	"igaku/encounter-service/models"
	commonsModels "igaku/commons/models"
)

func MigrateSchema(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Organization{},
		&commonsModels.Setting{},
	)

	return err
}
