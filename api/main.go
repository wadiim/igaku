package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	docs "igaku/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"igaku/controllers"
	"igaku/models"
	"igaku/repositories"
	"igaku/services"
)

// @title		Igaku API
// @version		0.0.1
// @host		localhost:8080

func main() {
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		os.Exit(1)
	}

	err = db.AutoMigrate(
		&models.Organization{},
		&models.Setting{},
	)
	if err != nil {
		log.Fatalf("Failed to create database structures: %v", err)
		os.Exit(1)
	}

	err = seedDatabase(db)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	helloController := controllers.NewHelloController()
	helloController.RegisterHelloRoutes(router)

	orgRepo := repositories.NewGormOrganizationRepository(db)
	orgService := services.NewOrganizationService(orgRepo)
	orgController := controllers.NewOrganizationController(orgService)
	orgController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}

func seedDatabase(db *gorm.DB) error {
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
