package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"log"

	"igaku/auth-service/clients"
	"igaku/auth-service/controllers"
	"igaku/auth-service/docs"
	"igaku/auth-service/services"
)

// @title		Igaku Auth API
// @version		0.0.1
// @host		localhost:8081

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	// TODO: Read URI from `.env`
	userClient, err := clients.NewUserClient("amqp://rabbit:tibbar@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect create a client: %v", err)
	}
	defer userClient.Shutdown()

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	authService := services.NewAuthService(userClient)
	authController := controllers.NewAuthController(authService)
	authController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}
