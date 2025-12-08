package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"igaku/commons/dtos"
	"igaku/geo-service/errors"
)

type GeoService interface {
	Search(address string) ([]dtos.Location, error)
	Reverse(lat, lon string) (*dtos.Location, error)
}

type geoService struct {
	nominatimURL string
	nominatimTimeout time.Duration
}

func NewGeoService() GeoService {
	nominatimURL := os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}

	t, err := strconv.Atoi(os.Getenv("NOMINATIM_TIMEOUT"))
	if err != nil || t <= 0 {
		t = 10
	}
	timeout := time.Duration(t) * time.Second

	return &geoService{
		nominatimURL: nominatimURL,
		nominatimTimeout: timeout,
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
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to create request",
		}
	}

	req.Header.Set("User-Agent", "curl/8.17.0")
	req.Header.Set("Accept", "*/*")

	ctx, cancel := context.WithTimeout(
		context.Background(), s.nominatimTimeout,
	)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			log.Printf("Request to external API timed out: %v\n", err)
			return nil, &errors.TimeoutError{}
		}
		log.Printf("Failed to connect to external API: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to perform a lookup",
		}
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf(
			"Request to external API failed with status code: %d\n",
			res.StatusCode,
		)
		if res.StatusCode == 400 {
			return nil, &errors.InvalidAddressError{}
		}
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to perform a lookup",
		}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to read response body",
		}
	}

	var locations []dtos.Location
	if err := json.Unmarshal(body, &locations); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
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
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to create request",
		}
	}

	req.Header.Set("User-Agent", "curl/8.17.0")
	req.Header.Set("Accept", "*/*")

	ctx, cancel := context.WithTimeout(
		context.Background(), s.nominatimTimeout,
	)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
			log.Printf("Request to external API timed out: %v\n", err)
			return nil, &errors.TimeoutError{}
		}
		log.Printf("Failed to connect to external API: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to perform a lookup",
		}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to read response body",
		}
	}

	var errMap map[string]interface{}
	if json.Unmarshal(body, &errMap) == nil {
		if errBody, ok := errMap["error"].(map[string]interface{}); ok {
			if val, ok := errBody["message"].(string); ok {
				msg := fmt.Sprintf("%v", val)
				ret := &errors.ExternalApiRequestError{
					Message: msg,
				}
				return nil, ret
			}
		} else if val, ok := errMap["error"].(string); ok {
			msg := fmt.Sprintf("%v", val)
			return nil, &errors.ExternalApiRequestError{Message: msg}
		}
	}

	if res.StatusCode != 200 {
		log.Printf(
			"Request to external API failed with status code: %d\n",
			res.StatusCode,
		)
		if res.StatusCode == 400 {
			return nil, &errors.InvalidAddressError{}
		}
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to perform a lookup",
		}
	}

	var location dtos.Location
	if err := json.Unmarshal(body, &location); err != nil {
		log.Printf("Failed to parse JSON: %v\n", err)
		return nil, &errors.ExternalApiRequestError{
			Message: "Failed to parse external API response",
		}
	}

	return &location, nil
}
