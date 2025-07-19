package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"igaku/user-service/repositories"
	"igaku/user-service/servers"
	"igaku/user-service/services"
	"igaku/user-service/utils"
)

// @title		Igaku User API
// @version		0.0.1
// @host		localhost:4000

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	db, err := utils.InitDatabase()
	if err != nil {
		log.Fatalf("%v", err)
	}

	userRepo := repositories.NewGormUserRepository(db)
	accService := services.NewAccountService(userRepo)

	amqpURI := os.Getenv("RABBITMQ_URL")
	rbServer, err := servers.NewRabbitMQServer(amqpURI, accService)
	failOnError(err, "[RabbitMQ] Failed to initialize server")
	defer rbServer.Shutdown()

	err = rbServer.Start()
	failOnError(err, "[RabbitMQ] Failed to start listeners")

	apiServer := servers.NewApiServer(accService)
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
