package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"igaku/commons/dtos"
	"igaku/geo-service/controllers"
	"igaku/geo-service/errors"
	"igaku/geo-service/tests/mocks"
)

func setupGeoRouter(t *testing.T, mockGeoService *mocks.GeoService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	geoController := controllers.NewGeoController(mockGeoService)

	router := gin.Default()
	geoController.RegisterRoutes(router)
	return router
}

func TestGeoController_Search_Success(t *testing.T) {
	mockService := new(mocks.GeoService)
	router := setupGeoRouter(t, mockService)

	address := "fire-dori-st"
	expectedLocations := []dtos.Location{
		{
			ID:   243772084,
			Lat:  "35.6656280",
			Lon:  "139.7016220",
			Name: "ファイアー通り, 神南一丁目, 神南, 渋谷区, 東京都, 150-0041, 日本",
		},
	}
	mockService.On("Search", address).Return(expectedLocations, nil)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/geo/search/%s", address),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var returnedLocations []dtos.Location
	err = json.Unmarshal(rec.Body.Bytes(), &returnedLocations)
	assert.NoError(t, err)
	assert.Equal(t, expectedLocations, returnedLocations)
	mockService.AssertExpectations(t)
}

func TestGeoController_Search_MultipleResults(t *testing.T) {
	mockService := new(mocks.GeoService)
	router := setupGeoRouter(t, mockService)

	address := "new-york"
	expectedLocations := []dtos.Location{
		{
			ID:   423955799,
			Lat:  "40.7127281",
			Lon:  "-74.0060152",
			Name: "City of New York, New York, United States",
		},
		{
			ID:   326302990,
			Lat:  "43.1561681",
			Lon:  "-75.8449946",
			Name: "New York, United States",
		},
	}
	mockService.On("Search", address).Return(expectedLocations, nil)

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/geo/search/%s", address),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var returnedLocations []dtos.Location
	err = json.Unmarshal(rec.Body.Bytes(), &returnedLocations)
	assert.NoError(t, err)
	assert.Equal(t, expectedLocations, returnedLocations)
	mockService.AssertExpectations(t)
}

func TestGeoController_Search_InvalidAddress(t *testing.T) {
	mockService := new(mocks.GeoService)
	router := setupGeoRouter(t, mockService)

	address := " "
	mockService.
		On("Search", url.QueryEscape(address)).
		Return(nil, &errors.InvalidAddressError{}).
		Once()

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/geo/search/%s", url.QueryEscape(address)),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertExpectations(t)
}
