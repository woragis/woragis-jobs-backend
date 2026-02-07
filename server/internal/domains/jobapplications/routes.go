package jobapplications

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"woragis-jobs-service/internal/domains/jobapplications/contracttypes"
	"woragis-jobs-service/internal/domains/jobapplications/dailyobjectives"
	"woragis-jobs-service/internal/domains/jobapplications/interviewstages"
	"woragis-jobs-service/internal/domains/jobapplications/joblevels"
	"woragis-jobs-service/internal/domains/jobapplications/responses"
)

// SetupRoutes registers job application endpoints and subdomain routes.
func SetupRoutes(api fiber.Router, handler Handler, responseHandler responses.Handler, stageHandler interviewstages.Handler, db *gorm.DB, logger *slog.Logger) {
	// Main job application list and creation routes (must come before /:id routes)
	api.Post("/", handler.CreateJobApplication)
	api.Get("/", handler.ListJobApplications)
	
	// Reference data routes (job levels and contract types) - must come before /:id routes
	// so that /job-levels matches the specific handler instead of being treated as an :id parameter
	joblevels.RegisterRoutes(api, db, logger)
	contracttypes.RegisterRoutes(api, db, logger)
	
	// Daily objectives and progress tracking - must come before /:id routes
	dailyobjectives.RegisterRoutes(api, db, logger)
	
	// Dynamic ID-based routes (must come last to avoid matching static paths)
	api.Get("/:id", handler.GetJobApplication)
	api.Patch("/:id/status", handler.UpdateJobApplicationStatus)
	api.Patch("/:id", handler.UpdateJobApplication)
	api.Delete("/:id", handler.DeleteJobApplication)
	api.Post("/:id/generate-cover-letter", handler.GenerateCoverLetter)
	
	// Subdomain routes with dynamic application ID
	responses.SetupRoutes(api.Group("/:applicationId/responses"), responseHandler)
	interviewstages.SetupRoutes(api.Group("/:applicationId/interview-stages"), stageHandler)
}

