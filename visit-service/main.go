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

	"igaku/visit-service/clients"
	"igaku/visit-service/controllers"
	"igaku/visit-service/docs"
	"igaku/visit-service/repositories"
	"igaku/visit-service/services"
	"igaku/visit-service/utils"
	commonsUtils "igaku/commons/utils"
)

// @title		Igaku Visit API
// @version		0.0.1
// @host		localhost:4000

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

	err = commonsUtils.SeedDatabase(db, "./visit-service/resources")
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	amqpURI := os.Getenv("RABBITMQ_URL")

	geoClient, err := clients.NewGeoClient(amqpURI)
	if err != nil {
		log.Fatalf("Failed to create a geo client: %v", err)
	}
	defer geoClient.Shutdown()

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	orgRepo := repositories.NewGormOrganizationRepository(db)
	orgService := services.NewOrganizationService(orgRepo)
	orgController := controllers.NewOrganizationController(orgService)
	orgController.RegisterRoutes(router)

	router.GET(
		"/visit/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)
	router.Run()
}
