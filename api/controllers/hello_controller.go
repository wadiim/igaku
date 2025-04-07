package controllers

import (
	"github.com/gin-gonic/gin"

	"net/http"

	"igaku/dto"
)

type HelloController struct {}

func NewHelloController() *HelloController {
	return &HelloController{}
}

// @Summary	Show a hello message
// @Description	Returns a static hello world message as a JSON object.
// @Tags	Hello
// @Produce	json
// @Success	200 {object} dto.HelloOutput
// @Router	/hello [get]
func (ctrl *HelloController) SayHello(c *gin.Context) {
	response := dto.HelloOutput{Message: "Hello world!"}
	c.JSON(http.StatusOK, response)
}

func (ctrl *HelloController) RegisterHelloRoutes(router *gin.Engine) {
	helloRoutes := router.Group("/hello")
	{
		helloRoutes.GET("", ctrl.SayHello)
	}
}
