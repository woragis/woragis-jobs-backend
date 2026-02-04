package jobapplications

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	
	"woragis-jobs-service/internal/domains/jobapplications/contracttypes"
	"woragis-jobs-service/internal/domains/jobapplications/interviewstages"
	"woragis-jobs-service/internal/domains/jobapplications/joblevels"
	"woragis-jobs-service/internal/domains/jobapplications/responses"
)

// SetupRoutes registers job application endpoints and subdomain routes.
func SetupRoutes(api fiber.Router, handler Handler, responseHandler responses.Handler, stageHandler interviewstages.Handler, db *gorm.DB, logger *slog.Logger) {
	// Main job application routes
	api.Post("/", handler.CreateJobApplication)
	api.Get("/", handler.ListJobApplications)
	api.Get("/:id", handler.GetJobApplication)
	api.Patch("/:id/status", handler.UpdateJobApplicationStatus)
	api.Patch("/:id", handler.UpdateJobApplication)
	api.Delete("/:id", handler.DeleteJobApplication)
	api.Post("/:id/generate-cover-letter", handler.GenerateCoverLetter)
	
	// Reference data routes (job levels and contract types)
	joblevels.RegisterRoutes(api, db, logger)
	contracttypes.RegisterRoutes(api, db, logger)
	
	// Subdomain routes
	responses.SetupRoutes(api.Group("/:applicationId/responses"), responseHandler)
	interviewstages.SetupRoutes(api.Group("/:applicationId/interview-stages"), stageHandler)
}

