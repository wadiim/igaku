package dtos

// Intended to be returned only to admin.
type AccountDetailsWithID struct {
	ID		string `json:"id" example:"0b6f13da-efb9-4221-9e89-e2729ae90030"`
	Username	string `json:"username" example:"jdoe"`
	Role		string `json:"role" example:"patient"`
}
