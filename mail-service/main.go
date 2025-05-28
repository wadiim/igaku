package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"igaku/mail-service/servers"
	"igaku/mail-service/services"
)

// @title		Igaku Mail API
// @version		0.0.1
// @host		localhost:8083

func main() {
	service := services.NewMailService()

	amqpURI := os.Getenv("RABBITMQ_URL")
	rbServer, err := servers.NewRabbitMQServer(amqpURI, service)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ server: %w", err)
	}

	err = rbServer.Start()
	if err != nil {
		log.Fatalf("Failed to start RabbitMQ listeners: %w", err)
	}
	defer rbServer.Shutdown()

	apiServer := servers.NewApiServer()
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
