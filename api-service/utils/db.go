package utils

import (
	"gorm.io/gorm"

	"fmt"

	"igaku/api-service/models"
	commonsModels "igaku/commons/models"
)

func MigrateSchema(db *gorm.DB) error {
	db.Exec(fmt.Sprintf(
		"CREATE TYPE role AS ENUM('%s', '%s', '%s');",
		commonsModels.Patient,
		commonsModels.Doctor,
		commonsModels.Admin,
	))
	err := db.AutoMigrate(
		&models.Organization{},
		&models.Setting{},
		&commonsModels.User{},
	)

	return err
}
