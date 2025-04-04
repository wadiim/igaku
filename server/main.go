package main

import (
	"github.com/gin-gonic/gin"

	"igaku/controllers"
)

func main() {
	helloController := controllers.NewHelloController()

	router := gin.Default()

	helloController.RegisterHelloRoutes(router)

	router.Run()
}
