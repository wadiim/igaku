package dtos

import "igaku/commons/dtos"

type LocationWithType struct {
	dtos.Location
	Type string `json:"osm_type" binding:"required" example:"way"`
}

func (lwt LocationWithType) StripType() dtos.Location {
	return dtos.Location{
		ID: lwt.ID,
		Lat: lwt.Lat,
		Lon: lwt.Lon,
		Name: lwt.Name,
	}
}
