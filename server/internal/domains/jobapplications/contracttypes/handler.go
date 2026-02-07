package contracttypes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	responseutil "woragis-jobs-service/pkg/response"
)

// Handler handles HTTP requests for contract types.
type Handler struct {
	service Service
}

// NewHandler creates a new contract type handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ListTypes handles GET /contract-types
// @Summary List all contract types
// @Description Get all available contract types
// @Tags contract-types
// @Produce json
// @Success 200 {array} ContractType
// @Failure 500 {object} response.ErrorResponse
// @Router /contract-types [get]
func (h *Handler) ListTypes(c *fiber.Ctx) error {
	ctx := c.Context()

	types, err := h.service.ListTypes(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, types)
}

// GetType handles GET /contract-types/:id
// @Summary Get a contract type by ID
// @Description Get a specific contract type by its ID
// @Tags contract-types
// @Produce json
// @Param id path string true "Contract Type ID"
// @Success 200 {object} ContractType
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /contract-types/{id} [get]
func (h *Handler) GetType(c *fiber.Ctx) error {
	ctx := c.Context()
	typeID := c.Params("id")

	// Validate UUID format
	if _, err := uuid.Parse(typeID); err != nil {
		   return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, "invalid UUID format"))
	}

	contractType, err := h.service.GetType(ctx, typeID)
	if err != nil {
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, contractType)
}

// handleError processes domain errors into HTTP responses
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	// Check if it's a DomainError
	   // ...existing code...

	// Fallback for unknown errors
	return responseutil.Error(c, fiber.StatusInternalServerError, 0, fiber.Map{
		"error_code": "CTM999",
		"message":    "Internal server error",
	})
}

// isNotFoundError checks if an error is a not found error.
func isNotFoundError(err error) bool {
	return IsNotFoundError(err)
}
