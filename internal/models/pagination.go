package models

// Pagination represents pagination information for list responses
// @Description Pagination metadata for paginated API responses
type Pagination struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"20"`
	TotalPages int `json:"totalPages" example:"5"`
	TotalItems int `json:"totalItems" example:"97"`
}
