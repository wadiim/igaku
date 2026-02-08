package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"igaku/med-service/clients"
	"igaku/med-service/repositories"
	"igaku/med-service/services"
	"igaku/med-service/utils"
	"igaku/med-service/servers"
)

// @title		Igaku Med API
// @version		0.0.1
// @host		localhost:4000

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	rxNormAPI := utils.NewRxNormAPI()
	db, err := utils.InitDatabase(rxNormAPI)
	if err != nil {
		log.Fatalf("%v", err)
	}

	amqpURI := os.Getenv("RABBITMQ_URL")

	userClient, err := clients.NewUserClient(amqpURI)
	if err != nil {
		log.Fatalf("Failed to create a user client: %v", err)
	}
	defer userClient.Shutdown()

	rxClassAPI := utils.NewRxClassAPI()

	medRepo := repositories.NewGormMedRepository(db)
	medService := services.NewMedService(rxClassAPI, userClient, medRepo)

	rbServer, err := servers.NewRabbitMQServer(amqpURI, medService)
	failOnError(err, "[RabbitMQ] Failed to initialize server")
	defer rbServer.Shutdown()

	err = rbServer.Start()
	failOnError(err, "[RabbitMQ] Failed to start listeners")

	apiServer := servers.NewApiServer(medService)
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
