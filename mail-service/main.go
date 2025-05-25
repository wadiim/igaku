package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"igaku/mail-service/servers"
	"igaku/mail-service/services"
)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
