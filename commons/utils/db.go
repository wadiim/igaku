package utils

import (
	"gorm.io/gorm"

	"errors"
	"log"
	"os"
	"path/filepath"

	"igaku/commons/models"
	commonsErrors "igaku/commons/errors"
)

func SeedDatabase(db *gorm.DB, path string) error {
	var setting models.Setting
	err := db.Where("key = ?", "db_seeded").First(&setting).Error
	if err == nil {
		log.Println(
			"Database already seeded " +
			"(found 'db_seeded' setting). Skipping seeding.",
		)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking seeding status: %w", err)
		return &commonsErrors.DatabaseError{}
	}

	initScriptDir, err := filepath.Abs(path)
	if err != nil {
		log.Printf(
			"Failed to get absolute path for resources/: %w", err,
		)
		return &commonsErrors.DatabaseError{}
	}
	initScriptPath := filepath.Join(initScriptDir, "init.sql")
	sqlBytes, err := os.ReadFile(initScriptPath)
	if err != nil {
		log.Printf(
			"Failed to read DB init script: %s: %w",
			initScriptPath,
			err,
		)
		return &commonsErrors.DatabaseError{}
	}
	sqlScript := string(sqlBytes)
	tx := db.Exec(sqlScript)
	if tx.Error != nil {
		log.Printf("Failed to execute init script: %w", tx.Error)
		return &commonsErrors.DatabaseError{}
	}

	seedMarker := models.Setting{Key: "db_seeded", Value: "true"}
	if err := db.Create(&seedMarker).Error; err != nil {
		log.Printf("Failed to create 'db_seeded' marker: %v", err)
		return &commonsErrors.DatabaseError{}
	}

	return nil
}
