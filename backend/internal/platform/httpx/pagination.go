package httpx

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationMeta(page, limit, total int) PaginationMeta {
	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
