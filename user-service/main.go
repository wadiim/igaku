package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"fmt"
	"log"
	"os"

	"igaku/user-service/controllers"
	"igaku/user-service/docs"
	"igaku/user-service/repositories"
	"igaku/user-service/services"
	"igaku/user-service/utils"
	commonsUtils "igaku/commons/utils"
)

// @title		Igaku User API
// @version		0.0.1
// @host		localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	err = commonsUtils.SeedDatabase(db)
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	helloController := controllers.NewHelloController()
	helloController.RegisterRoutes(router)

	userRepo := repositories.NewGormUserRepository(db)
	accService := services.NewAccountService(userRepo)
	accController := controllers.NewAccountController(accService)
	accController.RegisterRoutes(router)

	internalAccController := controllers.NewInternalAccountController(accService)
	internalAccController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}
