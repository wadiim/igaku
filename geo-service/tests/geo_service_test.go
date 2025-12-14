package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"igaku/geo-service/dtos"
	"igaku/geo-service/services"
	commonDtos "igaku/commons/dtos"
)

var server *httptest.Server

func TestMain(m *testing.M) {
	server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			address := r.URL.Query().Get("q")

			if address == "yoyogi,tokyo" {
				w.Header().Set(
					"Content-Type", "application/json",
				)
				json.NewEncoder(w).Encode([]dtos.LocationWithType{
					{
						Location: commonDtos.Location{
							ID: 17054674,
							Lat: "35.6823040",
							Lon: "139.6917362",
							Name: "代々木, 渋谷区, 東京都, 151-0053, 日本",
						},
						Type: "relation",
					},
					{
						Location: commonDtos.Location{
							ID: 7093269751,
							Lat: "35.6835456",
							Lon: "139.7015260",
							Name: "代々木, 四谷角筈線, 代々木一丁目, 代々木, 渋谷区, 東京都, 151-0053, 日本",
						},
						Type: "node",
					},
					{
						Location: commonDtos.Location{
							ID: 2558121954,
							Lat: "35.6839514",
							Lon: "139.7020806",
							Name: "代々木, 代々木駅(北口), 代々木一丁目, 代々木, 渋谷区, 東京都, 151-0053,  日本",
						},
						Type: "node",
					},
				})
			}
		}),
	)
	defer server.Close()

	exitCode := m.Run()

	os.Exit(exitCode)
}

func TestGeoService_Search_FiltersOutRelationItems(t *testing.T) {
	service := services.NewGeoService(server.URL)

	locations, err := service.Search("yoyogi,tokyo")
	require.NoError(t, err)

	require.NotEmpty(t, locations)

	var ids []int64;
	for _, loc := range locations {
		ids = append(ids, loc.ID)
	}

	assert.Contains(t, ids, int64(7093269751))
	assert.Contains(t, ids, int64(2558121954))
	assert.NotContains(t, ids, int64(17054674))
}
