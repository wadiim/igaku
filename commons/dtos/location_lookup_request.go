package dtos

type LocationLookupRequest struct {
	ID int64 `json:"id" binding:"required" example:"90394480"`
}

