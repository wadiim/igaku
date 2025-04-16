package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"net/http"
	"errors"

	"igaku/services"
	"igaku/dtos"
)

type OrganizationController struct {
	service services.OrganizationService
}

func NewOrganizationController(service services.OrganizationService) *OrganizationController {
	return &OrganizationController{service: service}
}

// GetByID retrieves a specific organization by its UUID.
// @Summary	Get organization by ID
// @Description	Retrieves details for a specific organization using its UUID.
// @Tags	Organizations
// @Produce	json
// @Param	id path string true "Organization ID (UUIDv4 format)"
// @Success	200 {object} models.Organization "Successfully retrieved organization"
// @Failure	400 {object} dots.ErrorResponse "Bad Request - Invalid UUID format"
// @Failure	404 {object} dtos.ErrorResponse "Not Found - Organization not found"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error - Failed to retrieve organization"
// @Router	/organizations/{id} [get]
func (ctrl *OrganizationController) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Message: "Invalid UUID format",
		})
		return
	}

	org, err := ctrl.service.GetOrganizationByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Message: "Organization not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Message: "Failed to retrieve organization",
			})
		}
		return
	}

	c.JSON(http.StatusOK, org)
}

func (ctrl *OrganizationController) RegisterRoutes(router *gin.Engine) {
	router.GET("/organizations/:id", ctrl.GetByID)
}
