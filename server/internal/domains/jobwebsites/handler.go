package jobwebsites

import (
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"woragis-jobs-service/pkg/response"
)

// Handler exposes job website endpoints.
type Handler interface {
	CreateJobWebsite(c *fiber.Ctx) error
	GetJobWebsite(c *fiber.Ctx) error
	ListJobWebsites(c *fiber.Ctx) error
	UpdateJobWebsite(c *fiber.Ctx) error
	ResetCounter(c *fiber.Ctx) error
	DeleteJobWebsite(c *fiber.Ctx) error
}

type handler struct {
	service Service
	logger  *slog.Logger
}

// NewHandler constructs a job website handler.
func NewHandler(service Service, logger *slog.Logger) Handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

type createJobWebsitePayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	BaseURL     string `json:"baseUrl"`
	LoginURL    string `json:"loginUrl"`
	DailyLimit  int    `json:"dailyLimit"`
}

type updateJobWebsitePayload struct {
	DailyLimit  *int    `json:"dailyLimit,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	BaseURL     *string `json:"baseUrl,omitempty"`
	LoginURL    *string `json:"loginUrl,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
}

func (h *handler) CreateJobWebsite(c *fiber.Ctx) error {
	var payload createJobWebsitePayload
	if err := c.BodyParser(&payload); err != nil {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"reason": "invalid request payload",
		}))
	}

	website, err := h.service.CreateJobWebsite(
		c.Context(),
		payload.Name,
		payload.DisplayName,
		payload.BaseURL,
		payload.LoginURL,
		payload.DailyLimit,
	)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, website)
}

func (h *handler) GetJobWebsite(c *fiber.Ctx) error {
	websiteID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return h.handleError(c, NewDomainError(ErrCodeInvalidPayload, err, map[string]interface{}{
			"reason": "invalid website id",
		}))
	}

	website, err := h.service.GetJobWebsite(c.Context(), websiteID)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, website)
}

func (h *handler) ListJobWebsites(c *fiber.Ctx) error {
	enabledOnly := false
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			enabledOnly = enabled
		}
	}

	websites, err := h.service.ListJobWebsites(c.Context(), enabledOnly)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, websites)
}

// handleError processes domain errors into HTTP responses
func (h *handler) handleError(c *fiber.Ctx, err error) error {
	if domainErr, ok := AsDomainError(err); ok {
		return response.Error(c, domainErr.GetHTTPStatus(), 0, fiber.Map{
			"error_code": domainErr.Code,
			"message":    domainErr.Message,
			"details":    domainErr.Context,
		})
	}
	return response.Error(c, fiber.StatusInternalServerError, 0, fiber.Map{
		"error_code": "WEB999",
		"message":    "Internal server error",
	})
}

func (h *handler) UpdateJobWebsite(c *fiber.Ctx) error {
	websiteID, err := uuid.Parse(c.Params("id"))
	   if err != nil {
		   return response.Error(c, fiber.StatusBadRequest, 0, fiber.Map{
			   "error_code": ErrCodeInvalidPayload,
			   "message": "invalid website id",
		   })
	   }

	var payload updateJobWebsitePayload
	   if err := c.BodyParser(&payload); err != nil {
		   return response.Error(c, fiber.StatusBadRequest, 0, fiber.Map{
			   "error_code": ErrCodeInvalidPayload,
			   "message": "invalid request payload",
		   })
	   }

	updates := JobWebsiteUpdates(payload)

	website, err := h.service.UpdateJobWebsite(c.Context(), websiteID, updates)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, website)
}

func (h *handler) ResetCounter(c *fiber.Ctx) error {
	websiteID, err := uuid.Parse(c.Params("id"))
	   if err != nil {
		   return response.Error(c, fiber.StatusBadRequest, 0, fiber.Map{
			   "error_code": ErrCodeInvalidPayload,
			   "message": "invalid website id",
		   })
	   }

	if err := h.service.ResetCount(c.Context(), websiteID); err != nil {
		return h.handleError(c, err)
	}

	website, err := h.service.GetJobWebsite(c.Context(), websiteID)
	if err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, website)
}

func (h *handler) DeleteJobWebsite(c *fiber.Ctx) error {
	websiteID, err := uuid.Parse(c.Params("id"))
	   if err != nil {
		   return response.Error(c, fiber.StatusBadRequest, 0, fiber.Map{
			   "error_code": ErrCodeInvalidPayload,
			   "message": "invalid website id",
		   })
	   }

	if err := h.service.DeleteJobWebsite(c.Context(), websiteID); err != nil {
		return h.handleError(c, err)
	}

	return response.Success(c, fiber.StatusOK, fiber.Map{
		"message": "website deleted successfully",
	})
}

// ...existing code...

