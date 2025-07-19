package utils

import (
	"gorm.io/gorm"

	"igaku/encounter-service/models"
	commonsErrors "igaku/commons/errors"
	commonsModels "igaku/commons/models"
)

func MigrateSchema(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Organization{},
		&commonsModels.Setting{},
	)

	if err != nil {
		log.Printf("Failed to migrate DB schema: %w", err)
		return &commonsErrors.DatabaseError{}
	} else {
		return nil
	}
}
