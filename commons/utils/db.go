package utils

import (
	"gorm.io/gorm"

	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"igaku/commons/models"
)

func SeedDatabase(db *gorm.DB) error {
	var setting models.Setting
	err := db.Where("key = ?", "db_seeded").First(&setting).Error
	if err == nil {
		log.Println(
			"Database already seeded " +
			"(found 'db_seeded' setting). Skipping seeding.",
		)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("Error checking seeding status: %w", err)
	}

	initScriptDir, err := filepath.Abs("./resources")
	if err != nil {
		return fmt.Errorf(
			"Failed to get absolute path for resources/: %w", err,
		)
	}
	initScriptPath := filepath.Join(initScriptDir, "init.sql")
	sqlBytes, err := os.ReadFile(initScriptPath)
	if err != nil {
		return fmt.Errorf(
			"Failed to read db init script: %s: %w",
			initScriptPath,
			err,
		)
	}
	sqlScript := string(sqlBytes)
	tx := db.Exec(sqlScript)
	if tx.Error != nil {
		return fmt.Errorf("Failed to execute init script: %w", tx.Error)
	}

	seedMarker := models.Setting{Key: "db_seeded", Value: "true"}
	if err := db.Create(&seedMarker).Error; err != nil {
		return fmt.Errorf("Failed to create 'db_seeded' marker: %v", err)
	}

	return nil
}
