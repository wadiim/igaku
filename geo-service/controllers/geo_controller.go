package controllers

import (
	"github.com/gin-gonic/gin"

	"igaku/geo-service/dtos"
	commonsDtos "igaku/commons/dtos"

	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	NORMATIM_URL = "https://nominatim.openstreetmap.org"
)

type GeoController struct {
}

func NewGeoController() *GeoController {
	return &GeoController{}
}

// Search	returns a list of geographical locations corresponding to the given address.
// @Summary	Lookup a location from address
// @Description	Performs geocoding, i.e. conversion of the given textual description or addres into geographic coordinates.
// @Produce	json
// @Param	address path string true "Textual description or address"
// @Success	200 {object} []dtos.Location "Success"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error"
// @Router	/geo/search/{address} [get]
func (ctrl *GeoController) Search(c *gin.Context) {
	address := c.Param("address")
	escaped := url.QueryEscape(address)

	requestUrl := fmt.Sprintf(
		"%s/search?q=%s&format=json&addressdetails=1",
		NORMATIM_URL, escaped,
	)
	log.Printf("URL = %s\n", requestUrl)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to create request",
		})
		return
	}

	req.Header.Set("User-Agent", "curl/8.17.0")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil || res.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to perform a lookup",
		})
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to read response body",
		})
		return
	}

	var locations []dtos.Location
	if err := json.Unmarshal(body, &locations); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to parse external API response",
		})
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
