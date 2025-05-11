package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"fmt"
	"os"
	"log"

	"igaku/encounter-service/controllers"
	"igaku/encounter-service/docs"
	"igaku/encounter-service/repositories"
	"igaku/encounter-service/services"
	"igaku/encounter-service/utils"
	commonsUtils "igaku/commons/utils"
)

// @title		Igaku Encounter API
// @version		0.0.1
// @host		localhost:8082

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
	}

	err = utils.MigrateSchema(db)
	if err != nil {
		log.Fatalf("Failed to create database structures: %v", err)
	}

	err = commonsUtils.SeedDatabase(db, "./encounter-service/resources")
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	orgRepo := repositories.NewGormOrganizationRepository(db)
	orgService := services.NewOrganizationService(orgRepo)
	orgController := controllers.NewOrganizationController(orgService)
	orgController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}
