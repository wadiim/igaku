package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	medErrors "igaku/med-service/errors"
	"igaku/med-service/middleware"
	"igaku/med-service/services"
)

type DrugController struct {
	service services.DrugService
}

func NewDrugController(service services.DrugService) *DrugController {
	return &DrugController{service: service}
}

func (ctrl *DrugController) GetRecommendedDrugs(c *gin.Context) {
	disease := c.Param("disease")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	orderByStr := c.DefaultQuery("orderBy", "id")
	orderMethodStr := c.DefaultQuery("orderMethod", "asc")

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

	orderBy, ok := commonsModels.DrugOrderableFieldsMap[strings.ToLower(orderByStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderBy parameter. Must be `id`, `name` or `substance`",
		})
		return
	}

	orderMethod, ok := commonsUtils.OrderingsMap[strings.ToLower(orderMethodStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderMethod parameter. Must be `asc` or `desc`",
		})
		return
	}
	log.Printf("%v", orderBy)
	log.Printf("%v", orderMethod)

	drugs, err := ctrl.service.GetRecommendedDrugs(disease, page, pageSize, orderBy, orderMethod)
	if err != nil {
		if errors.Is(err, &medErrors.RxClassUnavailableError{}) {
			c.JSON(http.StatusServiceUnavailable, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else if errors.Is(err, &medErrors.SubstanceNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve list of drugs",
			})
			return
		}
	}

	c.JSON(http.StatusOK, drugs)
}

func (ctrl *DrugController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/med/drug")
	routes.Use(middleware.Authenticate())
	{
		routes.GET(
			"/recommend/:disease",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetRecommendedDrugs,
		)
	}
}
