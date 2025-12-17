package main

import (
	// "github.com/gin-gonic/gin"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
	// actuator "github.com/sinhashubham95/go-actuator"

	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "igaku/med-service/controllers"
	// "igaku/med-service/docs"
	"igaku/med-service/repositories"
	"igaku/med-service/services"
	"igaku/med-service/utils"
	"igaku/med-service/servers"
	// configs "igaku/commons/configs"
)

// @title		Igaku Med API
// @version		0.0.1
// @host		localhost:4000

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	rxNormURL := "https://rxnav.nlm.nih.gov/REST/rxclass/allClasses?classTypes=DISEASE"
	rxAPI := utils.RxNormAPI{URL: rxNormURL}
	db, err := utils.InitDatabase(&rxAPI)
	if err != nil {
		log.Fatalf("%v", err)
	}

	diseaseRepo := repositories.NewGormDiseaseRepository(db)
	diseaseService := services.NewDiseaseService(diseaseRepo)

	patientRepo := repositories.NewGormPatientRepository(db)
	patientService := services.NewPatientService(patientRepo)

	amqpURI := os.Getenv("RABBITMQ_URL")
	rbServer, err := servers.NewRabbitMQServer(amqpURI, patientService)
	failOnError(err, "[RabbitMQ] Failed to initialize server")
	defer rbServer.Shutdown()

	err = rbServer.Start()
	failOnError(err, "[RabbitMQ] Failed to start listeners")

	apiServer := servers.NewApiServer(diseaseService)
	apiServer.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancelShutdown := context.WithTimeout(
		context.Background(), 10*time.Second,
	)
	defer cancelShutdown()

	if err = apiServer.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shutdown REST API")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
