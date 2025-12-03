package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"

	"igaku/med-service/services"
	commonsDtos "igaku/commons/dtos"
	igakuErrors "igaku/med-service/errors"
)

type DiseaseController struct {
	service services.DiseaseService
}

func NewDiseaseController(service services.DiseaseService) *DiseaseController {
	return &DiseaseController{service: service}
}

func (ctrl *DiseaseController) GetByName(c *gin.Context) {
	name := c.Param("name")

	diseases, err := ctrl.service.GetByName(name)
	if err != nil {
		if errors.Is(err, &igakuErrors.DiseaseNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve disease details",
			})
		}
		return
	}

	c.JSON(http.StatusOK, diseases)
	// TODO: Create repository in `main.go`
}

func (ctrl *DiseaseController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/med/disease")
	{
		routes.GET("/:name", ctrl.GetByName)
	}
}

