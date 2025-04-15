package utils

import (
	"gorm.io/gorm"

	"fmt"

	"igaku/models"
)

func MigrateSchema(db *gorm.DB) error {
	db.Exec(fmt.Sprintf(
		"CREATE TYPE role AS ENUM('%s', '%s', '%s');",
		models.Patient,
		models.Doctor,
		models.Admin,
	))
	err := db.AutoMigrate(
		&models.Organization{},
		&models.Setting{},
		&models.User{},
	)

	return err
}
