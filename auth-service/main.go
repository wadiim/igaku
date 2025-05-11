package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"igaku/auth-service/clients"
	"igaku/auth-service/controllers"
	"igaku/auth-service/docs"
	"igaku/auth-service/services"
)

// @title		Igaku Auth
// @version		0.0.1
// @host		localhost:8081

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	userClient := clients.NewUserClient("http://igaku-user:8080")
	authService := services.NewAuthService(userClient)
	authController := controllers.NewAuthController(authService)
	authController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run()
}
