package dtos

type Location struct {
	ID	int64	`json:"osm_id" binding:"required" example:"90394480"`
	Lat	string	`json:"lat" binding:"required" example:"52.5487921"`
	Lon	string 	`json:"lon" binding:"required" example:"-1.8164308"`
	Name	string 	`json:"display_name" binding:"required" example:"135, Pilkington Avenue, Maney, Sutton Coldfield, Birmingham, West Midlands, England, B72 1LH, United Kingdom"`
}
