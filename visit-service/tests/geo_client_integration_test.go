//go:build integration

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"igaku/commons/dtos"
	"igaku/visit-service/clients"
	testUtils "igaku/visit-service/tests/utils"
)

const amqpURL = "amqp://rabbit:tibbar@localhost:5672/"

var mockNominatim *httptest.Server

var fireDoriLoc = dtos.Location{
	ID:   153108796,
	Lat:  "35.6656280",
	Lon:  "139.7016220",
	Name: "ファイアー通り, 神南一丁目, 神南, 渋谷区, 東京都, 150-0041, 日本",
}

func TestMain(m *testing.M) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/reverse" {
			lat := r.URL.Query().Get("lat")
			lon := r.URL.Query().Get("lon")

			if lat == fireDoriLoc.Lat && lon == fireDoriLoc.Lon {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(fireDoriLoc)
			} else if lat == "0.0" && lon == "0.0" {
				time.Sleep(4 * time.Second)
				w.WriteHeader(http.StatusOK)
			} else if lat == "foo" && lon == "bar" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("{\"error\":{\"code\":400,\"message\":\"Parameter 'lon' must be a number.\"}}"))
			} else if lat == "90" && lon == "90" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("{\"error\":\"Unable to geocode\"}"))
			} else {
				http.NotFound(w, r)
			}
		} else if path == "/lookup" {
			ids := r.URL.Query().Get("osm_ids")

			if ids == fmt.Sprintf("N%d,W%d", fireDoriLoc.ID, fireDoriLoc.ID) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]dtos.Location{
					fireDoriLoc,
				})
			} else if ids == "N408,W408" {
				time.Sleep(4 * time.Second)
				w.WriteHeader(http.StatusOK)
			} else if ids == "N0,W0" {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("[]"))
			} else {
				http.NotFound(w, r)
			}
		} else {
			http.NotFound(w, r)
		}
	}))

	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}

	server.Listener = l
	server.Start()
	mockNominatim = server
	defer mockNominatim.Close()

	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancelCtx()

	u, err := url.Parse(mockNominatim.URL)
	if err != nil {
		log.Fatalf("Failed to parse URL: %v", err)
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		log.Fatalf("Failed to split URL: %v", err)
	}
	mockUrl := fmt.Sprintf("http://host.docker.internal:%s", port)

	log.Printf("mockUrl: %s", mockUrl)

	cleanup, err := testUtils.SetupTestServices(ctx, mockUrl)
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
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	lat := fireDoriLoc.Lat
	lon := fireDoriLoc.Lon
	expectedLocation := &fireDoriLoc
	location, err := geoClient.ReverseGeocode(lat, lon)
	require.NoError(t, err)
	assert.Equal(t, expectedLocation.ID, location.ID)
	assert.Equal(t, expectedLocation.Lat, location.Lat)
	assert.Equal(t, expectedLocation.Lon, location.Lon)
	assert.Equal(t, expectedLocation.Name, location.Name)
}

func TestGeoClient_ReverseGeocode_Timeout(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	lat := "0.0"
	lon := "0.0"

	_, err = geoClient.ReverseGeocode(lat, lon)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "request timed out")
}

func TestGeoClient_ReverseGeocode_NonGeocodableLocation(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	lat := "90"
	lon := "90"

	_, err = geoClient.ReverseGeocode(lat, lon)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "unable to geocode")
}

func TestGeoClient_ReverseGeocode_InvalidParams(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	lat := "foo"
	lon := "bar"

	_, err = geoClient.ReverseGeocode(lat, lon)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "must be a number")
}

func TestGeoClient_LocationLookup_Success(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	id := fireDoriLoc.ID
	expectedLocation := &fireDoriLoc
	location, err := geoClient.LookupLocation(int64(id))
	require.NoError(t, err)
	assert.Equal(t, expectedLocation.ID, location.ID)
	assert.Equal(t, expectedLocation.Lat, location.Lat)
	assert.Equal(t, expectedLocation.Lon, location.Lon)
	assert.Equal(t, expectedLocation.Name, location.Name)
}

func TestGeoClient_LocationLookup_Timeout(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	id := 408

	_, err = geoClient.LookupLocation(int64(id))
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "request timed out")
}

func TestGeoClient_LocationLookup_NoResult(t *testing.T) {
	geoClient, err := clients.NewGeoClient(amqpURL)
	require.NoError(t, err)
	defer geoClient.Shutdown()

	id := 0

	ret, err := geoClient.LookupLocation(int64(id))
	require.NoError(t, err)
	assert.Nil(t, ret)
}
