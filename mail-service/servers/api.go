package servers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"context"
	"log"
	"net/http"
	"time"

	"igaku/mail-service/controllers"
	"igaku/mail-service/docs"
)

type ApiServer struct {
	router *gin.Engine
	server *http.Server
}

func NewApiServer() *ApiServer {
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

	healthController := controllers.NewHealthController()
	healthController.RegisterRoutes(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
