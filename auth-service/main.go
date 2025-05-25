package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"log"
	"time"

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

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8090"}, // Swagger UI origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           1*time.Hour,
	}))

	// TODO: Read URI from `.env`
	amqpURI := "amqp://rabbit:tibbar@rabbitmq:5672/"

	userClient, err := clients.NewUserClient(amqpURI)
	if err != nil {
		log.Fatalf("Failed to create a user client: %v", err)
	}
	defer userClient.Shutdown()

	mailClient, err := clients.NewMailClient(amqpURI)
	if err != nil {
		log.Fatalf("Failed to create a mail client: %v", err)
	}
	defer mailClient.Shutdown()

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	authService := services.NewAuthService(userClient, mailClient)
	authController := controllers.NewAuthController(authService)
	authController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}
