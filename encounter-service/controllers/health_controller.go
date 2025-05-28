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

// @Summary	Check health
// @Description	Returns an OK message
// @Tags	Health
// @Produce	text/plain
// @Success	200 {string} string
// @Router	/encounter/health [get]
func (ctrl *HealthController) GetHealth(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func (ctrl *HealthController) RegisterRoutes(router *gin.Engine) {
	router.GET("/encounter/health", ctrl.GetHealth)
}
