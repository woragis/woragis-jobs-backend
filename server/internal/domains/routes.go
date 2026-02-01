package jobs

import (
	"log/slog"
	"os"
	"strconv"

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
	
	// Initialize HTTP publisher for resume-generator (replaces RabbitMQ)
	resumeGeneratorURL := os.Getenv("RESUME_GENERATOR_URL")
	timeoutMs := 5000
	if v := os.Getenv("RESUME_GENERATOR_TIMEOUT_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil && t > 0 {
			timeoutMs = t
		}
	}

	var resumePublisher resumes.RabbitMQPublisher = resumes.NewNoOpPublisher(logger)
	if resumeGeneratorURL != "" {
		resumePublisher = resumes.NewHTTPPublisher(resumeGeneratorURL, timeoutMs, logger)
		logger.Info("HTTP publisher initialized for resume-generator", "url", resumeGeneratorURL)
	} else {
		logger.Warn("RESUME_GENERATOR_URL not provided, using NoOp publisher")
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
	
	// Initialize resume queue for background job processing
	var resumeQueue resumes.Queue
	if dbManager.GetRedis() != nil {
		resumeQueue = resumes.NewRedisQueue(dbManager.GetRedis())
		logger.Info("Resume queue initialized successfully with Redis")
	} else {
		logger.Warn("Redis connection not available, resume queue will be nil")
	}
	
	// Create adapter to bridge job application service to resume handler
	jobAppServiceAdapter := newJobApplicationServiceAdapter(jobAppService)
	
	// Use NewHandlerWithJobApplicationService to enable resume generation
	resumeHandler := resumes.NewHandlerWithJobApplicationService(resumeService, jobAppServiceAdapter, resumeQueue, "", logger)
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

// SetupPublicRoutes sets up public routes (no authentication required)
func SetupPublicRoutes(app fiber.Router, dbManager *database.Manager, logger *slog.Logger) {
	db := dbManager.GetPostgres()
	
	// Initialize repository for public resume access
	resumeRepo := resumes.NewGormRepository(db)
	
	// Initialize HTTP publisher for resume-generator
	resumeGeneratorURL := os.Getenv("RESUME_GENERATOR_URL")
	timeoutMs := 5000
	if v := os.Getenv("RESUME_GENERATOR_TIMEOUT_MS"); v != "" {
		if t, err := strconv.Atoi(v); err == nil && t > 0 {
			timeoutMs = t
		}
	}
	
	var resumePublisher resumes.RabbitMQPublisher = resumes.NewNoOpPublisher(logger)
	if resumeGeneratorURL != "" {
		resumePublisher = resumes.NewHTTPPublisher(resumeGeneratorURL, timeoutMs, logger)
	}
	
	// Initialize service and handler for public routes
	resumeService := resumes.NewService(resumeRepo, resumePublisher, logger)
	
	var resumeQueue resumes.Queue
	if dbManager.GetRedis() != nil {
		resumeQueue = resumes.NewRedisQueue(dbManager.GetRedis())
	}
	
	jobAppServiceAdapter := newJobApplicationServiceAdapter(nil)
	resumeHandler := resumes.NewHandlerWithJobApplicationService(resumeService, jobAppServiceAdapter, resumeQueue, "", logger)
	
	// Setup public routes
	resumes.SetupPublicRoutes(app, resumeHandler)
}
