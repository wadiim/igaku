package dtos

type PaginatedResponse struct {
	Data		interface{}	`json:"data"`
	Page		int		`json:"page"`
	PageSize	int		`json:"page_size"`
	TotalPages	int		`json:"total_pages"`
	TotalCount	int64		`json:"total_count"`
}

