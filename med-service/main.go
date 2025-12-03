package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	actuator "github.com/sinhashubham95/go-actuator"

	"log"

	"igaku/med-service/controllers"
	"igaku/med-service/docs"
	"igaku/med-service/repositories"
	"igaku/med-service/services"
	"igaku/med-service/utils"
	configs "igaku/commons/configs"
)

// @title		Igaku Med API
// @version		0.0.1
// @host		localhost:4000

func main() {
	rxNormURL := "https://rxnav.nlm.nih.gov/REST/rxclass/allClasses?classTypes=DISEASE"
	api := utils.RxNormAPI{rxNormURL}
	db, err := utils.InitDatabase(&api)
	if err != nil {
		log.Fatalf("%v", err)
	}

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

	diseaseRepo := repositories.NewGormDiseaseRepository(db)
	diseaseService := services.NewDiseaseService(diseaseRepo)
	diseaseController := controllers.NewDiseaseController(diseaseService)
	diseaseController.RegisterRoutes(router)

	router.Run()

}
