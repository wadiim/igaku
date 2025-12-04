package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"igaku/commons/dtos"
	"igaku/geo-service/errors"
)

type GeoService interface {
	Search(address string) ([]dtos.Location, error)
	Reverse(lat, lon string) (*dtos.Location, error)
}

type geoService struct{
	nominatimURL string
}

func NewGeoService() GeoService {
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}

	return &geoService{
		nominatimURL: nominatimURL,
	}
}

func (s *geoService) Search(address string) ([]dtos.Location, error) {
	escaped := url.QueryEscape(address)

	requestUrl := fmt.Sprintf(
		"%s/search?q=%s&format=json",
		s.nominatimURL, escaped,
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

	if err != nil || (res != nil && res.StatusCode != 200) {
		log.Printf("Failed to connect to external API: %v\n", err)
		if res != nil && res.StatusCode == 400 {
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

func (s *geoService) Reverse(lat, lon string) (*dtos.Location, error) {
	requestUrl := fmt.Sprintf(
		"%s/reverse?lat=%s&lon=%s&format=json",
		s.nominatimURL, lat, lon,
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

	if err != nil || (res != nil && res.StatusCode != 200) {
		log.Printf("Failed to connect to external API: %v\n", err)
		if res != nil && res.StatusCode == 400 {
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

	var location dtos.Location
	if err := json.Unmarshal(body, &location); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return nil, &errors.RequestError{
			Message: "Failed to parse external API response",
		}
	}

	return &location, nil
}
