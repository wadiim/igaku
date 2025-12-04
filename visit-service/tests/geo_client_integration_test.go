//go:build integration

package tests

import (
	"context"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"igaku/commons/dtos"
	"igaku/visit-service/clients"
	testUtils "igaku/visit-service/tests/utils"
)

var mockNominatim *httptest.Server

func TestMain(m *testing.M) {
	// mockNominatim = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	lat := r.URL.Query().Get("lat")
	// 	lon := r.URL.Query().Get("lon")

	// 	if lat == "35.6656280" && lon == "139.7016220" {
	// 		w.Header().Set("Content-Type", "application/json")
	// 		json.NewEncoder(w).Encode(dtos.Location{
	// 			ID:   243772084,
	// 			Lat:  "35.6656280",
	// 			Lon:  "139.7016220",
	// 			Name: "ファイアー通り, 神南一丁目, 神南, 渋谷区, 東京都, 150-0041, 日本",
	// 		})
	// 	} else {
	// 		http.NotFound(w, r)
	// 	}
	// }))
	// defer mockNominatim.Close()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelCtx()

	// TODO: Fix Nominatim API mocking.
	// NOTE: The commented code does not work probably because the
	//       `geo` service cannot access the mocked API.
	// cleanup, err := testUtils.SetupTestServices(ctx, mockNominatim.URL)
	cleanup, err := testUtils.SetupTestServices(
		ctx,
		"https://nominatim.openstreetmap.org",
	)
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	exitCode := m.Run()

	if cleanup != nil {
		cleanup()
	}

	os.Exit(exitCode)
}

func TestGeoClient_ReverseGeocode_Success(t *testing.T) {
	url := "amqp://rabbit:tibbar@localhost:5672/"

	geoClient, err := clients.NewGeoClient(url)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	lat := "35.6656280"
	lon := "139.7016220"
	expectedLocation := &dtos.Location{
		ID:   242474988,
		Lat:  "35.6656280",
		Lon:  "139.7016220",
		Name: "ファイアー通り, 神南一丁目, 神南, 渋谷区, 東京都, 150-0041, 日本",
	}

	location, err := geoClient.ReverseGeocode(lat, lon)
	require.NoError(t, err)
	// NOTE: The returned ID is different each time for some reason.
	// assert.Equal(t, expectedLocation.ID, location.ID)
	assert.Equal(t, expectedLocation.Lat, location.Lat)
	assert.Equal(t, expectedLocation.Lon, location.Lon)
	assert.Equal(t, expectedLocation.Name, location.Name)
}

// TODO: Add more tests
