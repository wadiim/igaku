package dtos

// ReverseGeocodeRequest represents the request for reverse geocoding
type ReverseGeocodeRequest struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}
