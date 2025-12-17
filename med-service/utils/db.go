package utils

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"fmt"
	"log"
	"os"
	"time"

	"igaku/commons/models"
	commonsErrors "igaku/commons/errors"
	commonsUtils "igaku/commons/utils"
)

func MigrateSchema(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Disease{},
		&models.PatientRecord{},
		&models.Setting{},
	)
	if err != nil {
		log.Printf("Failed to migrate DB schema: %v", err)
		return &commonsErrors.DatabaseError{}
	} else {
		return nil
	}
}

func InitDatabase(api *RxNormAPI) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s "+
		"user=%s "+
		"password=%s "+
		"dbname=%s "+
		"port=5432 "+
		"sslmode=disable "+
		"TimeZone=Europe/Warsaw",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	prefixedWriter := &PrefixedWriter{
		Out:	os.Stdout,
		Prefix:	"[GORM] ",
	}

	prefixedLogger := PrefixedLogger{
		Interface: logger.New(
			log.New(prefixedWriter, "", log.LstdFlags),
			logger.Config{
				LogLevel:			logger.Info,
				IgnoreRecordNotFoundError:	false,
				Colorful:			false,
			},
		),
		Prefix: "[SQL] ",
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: prefixedLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, &commonsErrors.DatabaseError{}
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get the database object: %v", err)
		return nil, &commonsErrors.DatabaseError{}
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = MigrateSchema(db)
	if err != nil {
		log.Printf("Failed to create database structures: %v", err)
		return nil, &commonsErrors.DatabaseError{}
	}

	var count int64
	result := db.Find(&models.Disease{}).Count(&count)
	if result.Error != nil {
		log.Printf("Failed to query database: %v", result.Error)
		return nil, &commonsErrors.DatabaseError{}
	}
	if count == 0 {
		diseases := api.GetAllDiseases(db)
		result := db.Create(diseases)
		if result.Error != nil {
			log.Printf("%v", result.Error)
		}
	} else {
		log.Printf("Disease table present: fetch skipped")
	}

	err = commonsUtils.SeedDatabase(db, "./med-service/resources")
	if err != nil {
		log.Printf("Failed to seed database: %v", err)
		return nil, &commonsErrors.DatabaseError{}
	}

	return db, nil
}
