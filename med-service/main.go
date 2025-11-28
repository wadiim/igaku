package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	actuator "github.com/sinhashubham95/go-actuator"

	"igaku/med-service/controllers"
	"igaku/med-service/docs"
	configs "igaku/commons/configs"
)

// @title		Igaku Med API
// @version		0.0.1
// @host		localhost:4000

func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	router.GET(
		"/med/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	actuatorHandler := actuator.GetActuatorHandler(configs.ActuatorConfig)
	ginActuatorHandler := func(ctx *gin.Context) {
		actuatorHandler(ctx.Writer, ctx.Request)
	}

	router.GET(
		"/med/actuator/*endpoint",
		ginActuatorHandler,
	)

	router.Run()

}
