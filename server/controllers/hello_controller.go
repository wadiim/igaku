package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HelloController struct {}

func NewHelloController() *HelloController {
	return &HelloController{}
}

func (ctrl *HelloController) SayHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello world!",
	})
}

func (ctrl *HelloController) RegisterHelloRoutes(router *gin.Engine) {
	helloRoutes := router.Group("/hello")
	{
		helloRoutes.GET("", ctrl.SayHello)
	}
}
