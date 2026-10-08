package models

// ErrorResponse represents an error response from the API
// @Description Standard error response structure for all API errors
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
// @Description Detailed error information including code, message, and optional details
type ErrorDetail struct {
	Code    string   `json:"code" example:"VALIDATION_ERROR"`
	Message string   `json:"message" example:"Invalid input data"`
	Details []string `json:"details,omitempty" example:"firstName is required,email must be valid"`
}
