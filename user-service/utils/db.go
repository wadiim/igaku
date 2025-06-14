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
	commonsUtils "igaku/commons/utils"
)

func MigrateSchema(db *gorm.DB) error {
	db.Exec(fmt.Sprintf(
		"CREATE TYPE role AS ENUM('%s', '%s', '%s');",
		models.Patient,
		models.Doctor,
		models.Admin,
	))
	err := db.AutoMigrate(
		&models.Setting{},
		&models.User{},
	)

	return err
}

func InitDatabase() (*gorm.DB, error) {
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
		errmsg := fmt.Errorf(
			"Failed to connect to database: %v", err,
		)
		return nil, errmsg
	}

	sqlDB, err := db.DB()
	if err != nil {
		errmsg := fmt.Errorf(
			"Failed to get the database object: %v", err,
		)
		return nil, errmsg
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = MigrateSchema(db)
	if err != nil {
		errmsg := fmt.Errorf(
			"Failed to create database structures: %v", err,
		)
		return nil, errmsg
	}

	err = commonsUtils.SeedDatabase(db, "./user-service/resources")
	if err != nil {
		errmsg := fmt.Errorf(
			"Failed to seed database: %v", err,
		)
		return nil, errmsg
	}

	return db, nil
}
