package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"strconv"

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

func (ctrl *DiseaseController) GetBySubstring(c *gin.Context) {
	name := c.Param("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid page parameter. Must be a positive integer.",
		})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid pageSize parameter. Must be a positive integer.",
		})
		return
	}

	diseases, err := ctrl.service.GetBySubstring(name, page, pageSize)
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
}

func (ctrl *DiseaseController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/med/disease")
	{
		routes.GET(
			"/:name",
			ctrl.GetBySubstring,
		)
	}
}

