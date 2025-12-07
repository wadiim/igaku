package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"strconv"

	"igaku/med-service/middleware"
	"igaku/med-service/services"
	commonsDtos "igaku/commons/dtos"
	commonModels "igaku/commons/models"
	igakuErrors "igaku/med-service/errors"
)

type DiseaseController struct {
	service services.DiseaseService
}

func NewDiseaseController(service services.DiseaseService) *DiseaseController {
	return &DiseaseController{service: service}
}

// GetBySubstring retrieves the list of diseases that match provided substring
// @Summary	List diseases that match provided substring (Doctor)
// @Description	Retrieves a paginated list of diseases that match provided substring. Requires Doctor privileges.
// @Tags	Diseases
// @Produce	json
// @Param	name path string false "Disease name substring"
// @Param	page query int false "Page number (default: 1)" minimum(1)
// @Param	pageSize query int false "Number of items per page (default: 10)" minimum(1) maximum(100)
// @Success	200  {object}  dtos.PaginatedResponse{data=[]dtos.DiseaseDetails} "Successfully retrieved list of diseases"
// @Failure	400  {object}  dtos.ErrorResponse  "Bad Request - Invalid query parameters (name, page, pageSize)"
// @Failure	401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure	500  {object}  dtos.ErrorResponse  "Internal Server Error - Failed to retrieve diseases"
// @Security	BearerAuth
// @Router	/med/disease/{name} [get]
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
	routes.Use(middleware.Authenticate())
	{
		routes.GET(
			"/:name",
			middleware.Authorize(commonModels.Doctor),
			ctrl.GetBySubstring,
		)
	}
}

