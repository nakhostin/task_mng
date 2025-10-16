package response

type Meta struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
	Sort       string `json:"sort"`
}

func NewMeta(page, limit, total int, sort string) *Meta {
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	return &Meta{Page: page, Limit: limit, Total: total, TotalPages: totalPages, Sort: sort}
}
