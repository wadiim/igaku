package dtos

type Location struct {
	ID	int64	`json:"osm_id" binding:"required" example:"8398124"`
	Lat	string	`json:"lat" binding:"required" example:"40.7579554"`
	Lon	string 	`json:"lon" binding:"required" example:"-73.9855319"`
	Name	string 	`json:"display_name" binding:"required" example:"Manhattan, New York County, City of New York, New York, United States"`
}
