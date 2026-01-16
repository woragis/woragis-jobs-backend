package errors

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents the JSON error response structure
type ErrorResponse struct {
	Success bool                   `json:"success"`
	Error   ErrorDetail            `json:"error"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ErrorDetail contains the error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SendError sends a structured error response
func SendError(c *fiber.Ctx, err *AppError) error {
	statusCode := err.ToHTTPStatus()

	response := ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	}

	// Only include context in non-production or if explicitly needed
	if len(err.Context) > 0 {
		response.Context = err.Context
	}

	return c.Status(statusCode).JSON(response)
}

// HandleError is a convenience function to handle errors uniformly
func HandleError(c *fiber.Ctx, err error) error {
	// If it's already an AppError, send it
	if appErr, ok := err.(*AppError); ok {
		return SendError(c, appErr)
	}

	// Otherwise, wrap it as a generic server error
	appErr := Wrap(SERVER_INTERNAL_ERROR, err)
	return SendError(c, appErr)
}
