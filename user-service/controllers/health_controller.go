package controllers

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

type HealthController struct {
}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (ctrl *HealthController) GetHealth(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func (ctrl *HealthController) RegisterRoutes(router *gin.Engine) {
	router.GET("/user/health", ctrl.GetHealth)
}
