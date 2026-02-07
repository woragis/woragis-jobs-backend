package joblevels

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	responseutil "woragis-jobs-service/pkg/response"
)

// Handler handles HTTP requests for job levels.
type Handler struct {
	service Service
}

// NewHandler creates a new job level handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ListLevels handles GET /job-levels
// @Summary List all job levels
// @Description Get all available job levels with seniority and intensity combinations
// @Tags job-levels
// @Produce json
// @Success 200 {array} JobLevel
// @Failure 500 {object} response.ErrorResponse
// @Router /job-levels [get]
func (h *Handler) ListLevels(c *fiber.Ctx) error {
	ctx := c.Context()

	levels, err := h.service.ListLevels(ctx)
	if err != nil {
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, levels)
}

// GetLevel handles GET /job-levels/:id
// @Summary Get a job level by ID
// @Description Get a specific job level by its ID
// @Tags job-levels
// @Produce json
// @Param id path string true "Job Level ID"
// @Success 200 {object} JobLevel
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /job-levels/{id} [get]
func (h *Handler) GetLevel(c *fiber.Ctx) error {
	ctx := c.Context()
	levelID := c.Params("id")

	// Validate UUID format
	if _, err := uuid.Parse(levelID); err != nil {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"field":   "id",
			"reason":  "invalid UUID format",
		}))
	}

	level, err := h.service.GetLevel(ctx, levelID)
	if err != nil {
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, level)
}

// handleError processes domain errors into HTTP responses
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	// Check if it's a DomainError
	if domainErr, ok := AsDomainError(err); ok {
		return responseutil.Error(c, domainErr.GetHTTPStatus(), 0, fiber.Map{
			"error_code": domainErr.Code,
			"message":    domainErr.Message,
			"details":    domainErr.Context,
		})
	}

	// Fallback for unknown errors
	return responseutil.Error(c, fiber.StatusInternalServerError, 0, fiber.Map{
		"error_code": "JOB999",
		"message":    "Internal server error",
	})
}

// isNotFoundError checks if an error is a not found error.
func isNotFoundError(err error) bool {
	return IsNotFoundError(err)
}
