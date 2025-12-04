package dtos

type ReverseGeocodeRequest struct {
	Lat string `json:"lat" binding:"required" example:"40.7579554"`
	Lon string `json:"lon" binding:"required" example:"-73.9855319"`
}
