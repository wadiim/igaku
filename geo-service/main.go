package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"igaku/geo-service/controllers"
	"igaku/geo-service/docs"
)

// @title		Igaku Geo API
// @version		0.0.1
// @host		localhost:4000

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	router.GET(
		"/geo/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	router.Run()
}
