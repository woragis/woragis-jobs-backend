package dailyobjectives

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	responseutil "woragis-jobs-service/pkg/response"
)

// Handler handles HTTP requests for daily objectives.
type Handler struct {
	service Service
}

// NewHandler creates a new daily objectives handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateObjective handles POST /daily-objectives
// @Summary Create daily objectives
// @Description Create daily application targets for the authenticated user
// @Tags daily-objectives
// @Accept json
// @Produce json
// @Param request body CreateObjectiveRequest true "Objective targets"
// @Success 201 {object} DailyObjective
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /daily-objectives [post]
func (h *Handler) CreateObjective(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "missing or invalid authentication token",
		}))
	}

	var parsedUserID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		parsedUserID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeValidation, err, map[string]interface{}{
				"field":   "userID",
				"reason":  "invalid user ID format",
			}))
		}
		parsedUserID = id
	default:
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "invalid user ID type",
		}))
	}

	var req CreateObjectiveRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"reason": "invalid JSON body",
		}))
	}

	if req.TotalTarget == 0 && req.JuniorTarget == 0 && req.PlenoTarget == 0 && req.SeniorTarget == 0 {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"reason": "missing or invalid fields: totalTarget, juniorTarget, plenoTarget, seniorTarget must be numbers and not null",
		}))
	}

	objective, err := h.service.CreateObjective(c.Context(), parsedUserID, req)
	if err != nil {
		if IsValidationError(err) {
			return h.handleError(c, err)
		}
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusCreated, objective)
}

// handleError processes domain errors into HTTP responses
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
	if domainErr, ok := AsDomainError(err); ok {
		return responseutil.Error(c, domainErr.GetHTTPStatus(), 0, fiber.Map{
			"error_code": domainErr.Code,
			"message":    domainErr.Message,
			"details":    domainErr.Context,
		})
	}
	return responseutil.Error(c, fiber.StatusInternalServerError, 0, fiber.Map{
		"error_code": "OBJ999",
		"message":    "Internal server error",
	})
}

// GetObjective handles GET /daily-objectives
// @Summary Get user's daily objectives
// @Description Get the authenticated user's daily application targets
// @Tags daily-objectives
// @Produce json
// @Success 200 {object} DailyObjective
// @Failure 404 {object} response.ErrorResponse
// @Router /daily-objectives [get]
func (h *Handler) GetObjective(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "missing or invalid authentication token",
		}))
	}

	var parsedUserID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		parsedUserID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeValidation, err, map[string]interface{}{
				"field":   "userID",
				"reason":  "invalid user ID format",
			}))
		}
		parsedUserID = id
	default:
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "invalid user ID type",
		}))
	}

	objective, err := h.service.GetObjective(c.Context(), parsedUserID)
	if err != nil {
		if IsNotFoundError(err) {
			return h.handleError(c, err)
		}
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, objective)
}

// UpdateObjective handles PATCH /daily-objectives
// @Summary Update daily objectives
// @Description Update the authenticated user's daily application targets
// @Tags daily-objectives
// @Accept json
// @Produce json
// @Param request body CreateObjectiveRequest true "Updated objective targets"
// @Success 200 {object} DailyObjective
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /daily-objectives [patch]
func (h *Handler) UpdateObjective(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "missing or invalid authentication token",
		}))
	}

	var parsedUserID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		parsedUserID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeValidation, err, map[string]interface{}{
				"field":   "userID",
				"reason":  "invalid user ID format",
			}))
		}
		parsedUserID = id
	default:
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "invalid user ID type",
		}))
	}

	var req CreateObjectiveRequest
	if err := c.BodyParser(&req); err != nil {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"reason": "invalid JSON body",
		}))
	}

	objective, err := h.service.UpdateObjective(c.Context(), parsedUserID, req)
	if err != nil {
		if IsValidationError(err) {
			return h.handleError(c, err)
		}
		if IsNotFoundError(err) {
			return h.handleError(c, err)
		}
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, objective)
}

// GetTodayProgress handles GET /daily-progress/today
// @Summary Get today's progress
// @Description Get the authenticated user's progress for today against their objectives
// @Tags daily-objectives
// @Produce json
// @Success 200 {object} DailyProgress
// @Failure 404 {object} response.ErrorResponse
// @Router /daily-progress/today [get]
func (h *Handler) GetTodayProgress(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "missing or invalid authentication token",
		}))
	}

	var parsedUserID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		parsedUserID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeValidation, err, map[string]interface{}{
				"field":   "userID",
				"reason":  "invalid user ID format",
			}))
		}
		parsedUserID = id
	default:
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "invalid user ID type",
		}))
	}

	progress, err := h.service.GetTodayProgress(c.Context(), parsedUserID)
	if err != nil {
		if IsNotFoundError(err) {
			return h.handleError(c, err)
		}
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, progress)
}

// GetHistoricalProgress handles GET /daily-progress/history
// @Summary Get historical progress
// @Description Get the authenticated user's progress for a date range. Supports presets (7days, 30days, 90days) or custom dates.
// @Tags daily-objectives
// @Produce json
// @Param preset query string false "Preset: 7days, 30days, 90days"
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} DailyProgress
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /daily-progress/history [get]
func (h *Handler) GetHistoricalProgress(c *fiber.Ctx) error {
	userIDRaw := c.Locals("userID")
	if userIDRaw == nil {
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "missing or invalid authentication token",
		}))
	}

	var parsedUserID uuid.UUID
	switch v := userIDRaw.(type) {
	case uuid.UUID:
		parsedUserID = v
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeValidation, err, map[string]interface{}{
				"field":   "userID",
				"reason":  "invalid user ID format",
			}))
		}
		parsedUserID = id
	default:
		return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
			"field":   "userID",
			"reason":  "invalid user ID type",
		}))
	}

	preset := c.Query("preset", "7days")
	fromStr := c.Query("from")
	toStr := c.Query("to")

	var from, to time.Time

	if preset != "" && fromStr == "" && toStr == "" {
		// Use preset
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		switch preset {
		case "7days":
			from = today.AddDate(0, 0, -6)
			to = today
		case "30days":
			from = today.AddDate(0, 0, -29)
			to = today
		case "90days":
			from = today.AddDate(0, 0, -89)
			to = today
		default:
			return h.handleError(c, NewDomainError(ErrCodeValidation, nil, map[string]interface{}{
				"reason": "Invalid preset. Use 7days, 30days, or 90days",
			}))
		}
	} else {
		// Use custom dates
		if fromStr == "" || toStr == "" {
			return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, nil, map[string]interface{}{
				"reason": "Both from and to dates required for custom range",
			}))
		}

		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
				"reason": "Invalid from date format. Use YYYY-MM-DD",
			}))
		}
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
				"reason": "Invalid to date format. Use YYYY-MM-DD",
			}))
		}

		from = parsedFrom
		to = parsedTo
	}

	// Validate date range
	if from.After(to) {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, nil, map[string]interface{}{
			"reason": "from date cannot be after to date",
		}))
	}

	// Limit range to max 365 days
	maxDays := 365
	daysDiff := int(to.Sub(from).Hours() / 24)
	if daysDiff > maxDays {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, nil, map[string]interface{}{
			"reason": "Date range cannot exceed 365 days",
		}))
	}

	progress, err := h.service.GetHistoricalProgress(c.Context(), parsedUserID, from, to)
	if err != nil {
		if IsNotFoundError(err) {
			return h.handleError(c, err)
		}
		return h.handleError(c, err)
	}

	return responseutil.Success(c, fiber.StatusOK, progress)
}
