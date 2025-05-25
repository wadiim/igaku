package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"igaku/user-service/repositories"
	"igaku/user-service/server"
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

	err = commonsUtils.SeedDatabase(db, "./user-service/resources")
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	userRepo := repositories.NewGormUserRepository(db)
	accService := services.NewAccountService(userRepo)

	amqpURI := os.Getenv("RABBITMQ_URL")
	rbServer, err := server.NewRabbitMQServer(amqpURI, accService)
	failOnError(err, "Failed to initialize RabbitMQ server")
	defer rbServer.Shutdown()

	err = rbServer.Start()
	failOnError(err, "Failed to start RabbitMQ listeners")

	apiServer := server.NewApiServer(accService)
	apiServer.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancelShutdown()

	if err = apiServer.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shutdown REST API")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
