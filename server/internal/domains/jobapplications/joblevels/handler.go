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
		return responseutil.Error(c, fiber.StatusInternalServerError, 11100, err.Error())
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
		return responseutil.Error(c, fiber.StatusBadRequest, 11100, err.Error())
	}

	level, err := h.service.GetLevel(ctx, levelID)
	if err != nil {
		if isNotFoundError(err) {
			return responseutil.Error(c, fiber.StatusNotFound, 11104, err.Error())
		}
		return responseutil.Error(c, fiber.StatusInternalServerError, 11103, err.Error())
	}

	return responseutil.Success(c, fiber.StatusOK, level)
}

// isNotFoundError checks if an error is a not found error.
func isNotFoundError(err error) bool {
	return IsNotFoundError(err)
}
