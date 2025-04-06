package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"net/http"
	"errors"

	"igaku/services"
)

type OrganizationController struct {
	service services.OrganizationService
}

func NewOrganizationController(service services.OrganizationService) *OrganizationController {
	return &OrganizationController{service: service}
}

func (ctrl *OrganizationController) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid UUID format",
		})
		return
	}

	org, err := ctrl.service.GetOrganizationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Organization not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve organization",
			})
		}
		return
	}

	c.JSON(http.StatusOK, org)
}

func (ctrl *OrganizationController) RegisterRoutes(router *gin.Engine) {
	router.GET("/organizations/:id", ctrl.GetByID)
}
