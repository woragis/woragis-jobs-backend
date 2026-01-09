package jobs

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"woragis-jobs-service/internal/database"
	"woragis-jobs-service/internal/domains/jobapplications"
	"woragis-jobs-service/internal/domains/jobapplications/interviewstages"
	"woragis-jobs-service/internal/domains/jobapplications/responses"
	"woragis-jobs-service/internal/domains/jobwebsites"
	"woragis-jobs-service/internal/domains/resumes"
	"woragis-jobs-service/pkg/aiservice"
	authPkg "woragis-jobs-service/pkg/auth"
	"woragis-jobs-service/pkg/middleware"
)

// SetupRoutes sets up all jobs service routes
func SetupRoutes(api fiber.Router, dbManager *database.Manager, jwtManager *authPkg.JWTManager, aiServiceURL string, logger *slog.Logger) {
	db := dbManager.GetPostgres()
	// Apply JWT validation middleware to all routes (local validation, no HTTP calls)
	if jwtManager != nil {
		api.Use(middleware.JWTMiddleware(middleware.JWTConfig{
			JWTManager: jwtManager,
		}))
	}

	// Initialize repositories
	jobAppRepo := jobapplications.NewGormRepository(db)
	resumeRepo := resumes.NewGormRepository(db)
	jobWebsiteRepo := jobwebsites.NewGormRepository(db)

	// Initialize services
	jobAppService := jobapplications.NewService(jobAppRepo, nil, logger) // Queue will be nil for now
	
	// Initialize RabbitMQ publisher for resume jobs
	var resumePublisher resumes.RabbitMQPublisher = resumes.NewNoOpPublisher(logger)
	if dbManager.GetRabbitMQ() != nil {
		var err error
		resumePublisher, err = resumes.NewRabbitMQPublisher(dbManager.GetRabbitMQ().Channel, logger)
		if err != nil {
			logger.Warn("failed to initialize RabbitMQ publisher", "error", err)
			resumePublisher = resumes.NewNoOpPublisher(logger)
		} else {
			logger.Info("RabbitMQ publisher initialized successfully")
		}
	} else {
		logger.Warn("RabbitMQ connection not available, using no-op publisher")
	}
	
	resumeService := resumes.NewService(resumeRepo, resumePublisher, logger)
	jobWebsiteService := jobwebsites.NewService(jobWebsiteRepo, logger)

	// Initialize AI service client for cover letter generation
	var coverLetterGenerator jobapplications.CoverLetterGenerator
	if aiServiceURL != "" {
		aiClient := aiservice.NewClient(aiServiceURL)
		coverLetterGenerator = jobapplications.NewAIServiceCoverLetterGenerator(aiClient, logger)
		logger.Info("AI service client initialized for cover letter generation", "url", aiServiceURL)
	} else {
		logger.Warn("AI service URL not provided, cover letter generation will be disabled")
	}

	// Initialize handlers
	var jobAppHandler jobapplications.Handler
	if coverLetterGenerator != nil {
		jobAppHandler = jobapplications.NewHandlerWithDependencies(jobAppService, nil, nil, coverLetterGenerator, logger)
	} else {
		jobAppHandler = jobapplications.NewHandler(jobAppService, logger)
	}
	resumeHandler := resumes.NewHandler(resumeService, nil, "", logger) // Queue and baseFilePath will be nil/empty for now
	jobWebsiteHandler := jobwebsites.NewHandler(jobWebsiteService, logger)

	// Initialize subdomain handlers
	responseRepo := responses.NewGormRepository(db)
	responseService := responses.NewService(responseRepo, logger)
	responseHandler := responses.NewHandler(responseService, logger)
	
	stageRepo := interviewstages.NewGormRepository(db)
	stageService := interviewstages.NewService(stageRepo, logger)
	stageHandler := interviewstages.NewHandler(stageService, logger)

	// Setup routes
	jobapplications.SetupRoutes(api.Group("/job-applications"), jobAppHandler, responseHandler, stageHandler)
	resumes.SetupRoutes(api.Group("/resumes"), resumeHandler)
	jobwebsites.SetupRoutes(api.Group("/job-websites"), jobWebsiteHandler)
}
