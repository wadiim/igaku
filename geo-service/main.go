package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"igaku/geo-service/controllers"
	"igaku/geo-service/docs"
	"igaku/geo-service/servers"
	"igaku/geo-service/services"
)

// @title		Igaku Geo API
// @version		0.0.1
// @host		localhost:4000

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}

	geoService := services.NewGeoService(nominatimURL)
	geoController := controllers.NewGeoController(geoService)
	geoController.RegisterRoutes(router)

	amqpURI := os.Getenv("RABBITMQ_URL")

	rbServer, err := servers.NewRabbitMQServer(amqpURI, geoService)
	if err != nil {
		log.Fatalf("[RabbitMQ] Failed to start listeners: %v", err)
	}
	defer rbServer.Shutdown()

	if err := rbServer.Start(); err != nil {
		log.Fatalf("Failed to start RabbitMQ server: %v", err)
	}

	router.GET(
		"/geo/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	router.Run()
}
