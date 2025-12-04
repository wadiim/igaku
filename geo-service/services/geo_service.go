package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"igaku/geo-service/dtos"
	"igaku/geo-service/errors"
)

const (
	NORMATIM_URL = "https://nominatim.openstreetmap.org"
)

type GeoService interface {
	Search(address string) ([]dtos.Location, error)
}

type geoService struct {}

func NewGeoService() GeoService {
	return &geoService{}
}

func (s *geoService) Search(address string) ([]dtos.Location, error) {
	escaped := url.QueryEscape(address)

	requestUrl := fmt.Sprintf(
		"%s/search?q=%s&format=json",
		NORMATIM_URL, escaped,
	)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		log.Printf("Failed to create request: %v\n", err)
		return nil, &errors.RequestError{
			Message: "Failed to create request",
		}
	}

	req.Header.Set("User-Agent", "curl/8.17.0")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil || res.StatusCode != 200 {
		if res.StatusCode == 400 {
			return nil, &errors.InvalidAddressError{}
		}
		return nil, &errors.RequestError{
			Message: "Failed to perform a lookup",
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return nil, &errors.RequestError{
			Message: "Failed to read response body",
		}
	}

	var locations []dtos.Location
	if err := json.Unmarshal(body, &locations); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return nil, &errors.RequestError{
			Message: "Failed to parse external API response",
		}
	}

	return locations, nil
}
