package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"igaku/geo-service/services"
	commonsDtos "igaku/commons/dtos"
	igakuErrors "igaku/geo-service/errors"
)

type GeoController struct {
	service services.GeoService
}

func NewGeoController(service services.GeoService) *GeoController {
	return &GeoController{
		service: service,
	}
}

// Search	returns a list of geographical locations corresponding to the given address.
// @Summary	Lookup a location from address
// @Description	Performs geocoding, i.e. conversion of the given textual description or addres into geographic coordinates.
// @Produce	json
// @Param	address path string true "Textual description or address"
// @Success	200 {object} []dtos.Location "Success"
// @Failure	400 {object} dtos.ErrorResponse "Invalid Request"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error"
// @Router	/geo/search/{address} [get]
func (ctrl *GeoController) Search(c *gin.Context) {
	address := c.Param("address")
	locations, err := ctrl.service.Search(address)

	if err != nil {
		if errors.Is(err, &igakuErrors.InvalidAddressError{}) {
			c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to perform a lookup",
			})
		}
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (ctrl *GeoController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/geo")
	{
		routes.GET("/search/:address", ctrl.Search)
	}
}
