package domain

import "math"

// PaginationQuery holds query parameters for paginated and filtered list requests.
type PaginationQuery struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Status   string `form:"status"`
	FileCode string `form:"file_code"`
}

// Normalize ensures defaults: Page >= 1, Limit clamped to [1, 100].
func (p *PaginationQuery) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// Offset returns the SQL OFFSET value derived from Page and Limit.
func (p *PaginationQuery) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResponse wraps any list response with pagination metadata.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse constructs a PaginatedResponse from data, total count, and query.
func NewPaginatedResponse(data interface{}, total int64, q PaginationQuery) *PaginatedResponse {
	totalPages := int(math.Ceil(float64(total) / float64(q.Limit)))
	return &PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       q.Page,
		Limit:      q.Limit,
		TotalPages: totalPages,
	}
}
