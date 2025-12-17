package servers

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	actuator "github.com/sinhashubham95/go-actuator"

	"context"
	"log"
	"net/http"

	"igaku/med-service/controllers"
	"igaku/med-service/docs"
	"igaku/med-service/services"
	configs "igaku/commons/configs"
)


type ApiServer struct {
	router *gin.Engine
	server *http.Server
}

func NewApiServer(diseaseService services.DiseaseService) *ApiServer {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	diseaseController := controllers.NewDiseaseController(diseaseService)
	diseaseController.RegisterRoutes(router)

	actuatorHandler := actuator.GetActuatorHandler(configs.ActuatorConfig)
	ginActuatorHandler := func(ctx *gin.Context) {
		actuatorHandler(ctx.Writer, ctx.Request)
	}

	router.GET(
		"/med/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	router.GET(
		"/med/actuator/*endpoint",
		ginActuatorHandler,
	)

	server := &http.Server{
		Addr: ":8080",
		Handler: router,
	}

	return &ApiServer{router: router, server: server}
}

func (s *ApiServer) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to serve REST API: %v", err)
		}
	}()
}

func (s *ApiServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
