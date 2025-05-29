package controllers

import (
	"github.com/gin-gonic/gin"

	"net/http"

	"igaku/user-service/dtos"
)

type HelloController struct {}

func NewHelloController() *HelloController {
	return &HelloController{}
}

// @Summary	Show a hello message
// @Description	Returns a static hello world message as a JSON object.
// @Tags	Hello
// @Produce	json
// @Success	200 {object} dtos.HelloOutput
// @Router	/user/hello [get]
func (ctrl *HelloController) SayHello(c *gin.Context) {
	response := dtos.HelloOutput{Message: "Hello world!"}
	c.JSON(http.StatusOK, response)
}

func (ctrl *HelloController) RegisterRoutes(router *gin.Engine) {
	helloRoutes := router.Group("/user/hello")
	{
		helloRoutes.GET("", ctrl.SayHello)
	}
}
