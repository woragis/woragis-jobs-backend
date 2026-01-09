package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"woragis-jobs-service/internal/config"
	"woragis-jobs-service/internal/database"
	"woragis-jobs-service/pkg/health"
	applogger "woragis-jobs-service/pkg/logger"
	appmetrics "woragis-jobs-service/pkg/metrics"
	appsecurity "woragis-jobs-service/pkg/security"
	apptimeout "woragis-jobs-service/pkg/timeout"
	apptracing "woragis-jobs-service/pkg/tracing"

	jobsdomain "woragis-jobs-service/internal/domains"
	authPkg "woragis-jobs-service/pkg/auth"
)

func main() {
	// Load configuration first to get environment
	cfg := config.Load()
	env := cfg.Env
	if env == "" {
		env = os.Getenv("ENV")
		if env == "" {
			env = "development"
		}
	}

	// Setup structured logger with trace ID support
	slogLogger := applogger.New(env)

	// Initialize OpenTelemetry tracing
	tracingShutdown, err := apptracing.Init(apptracing.Config{
		ServiceName:    cfg.AppName,
		ServiceVersion: "1.0.0", // TODO: Get from build info
		Environment:    env,
		JaegerEndpoint: os.Getenv("JAEGER_ENDPOINT"), // Defaults to http://jaeger:4318
	})
	if err != nil {
		slogLogger.Warn("failed to initialize tracing", "error", err)
	} else {
		slogLogger.Info("tracing initialized", "service", cfg.AppName)
		defer func() {
			if tracingShutdown != nil {
				tracingShutdown()
			}
		}()
	}

	// Load database and Redis configs
	dbCfg := config.LoadDatabaseConfig()
	redisCfg := config.LoadRedisConfig()

	// Initialize database manager
	dbManager, err := database.NewFromConfig(dbCfg, redisCfg)
	if err != nil {
		slogLogger.Error("failed to initialize database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// Perform initial health check
	if err := dbManager.HealthCheck(); err != nil {
		slogLogger.Warn("Database health check failed", "error", err)
	} else {
		slogLogger.Info("All database connections are healthy")
	}

	// Run migrations
	if err := jobsdomain.MigrateJobsTables(dbManager.GetPostgres()); err != nil {
		slogLogger.Error("failed to run jobs migrations", "error", err)
		os.Exit(1)
	}

	// Create Fiber app
	app := config.CreateFiberApp(cfg)

	// Recovery middleware (early in chain)
	app.Use(recover.New())

	// Security headers middleware (must be early, before other middlewares)
	app.Use(appsecurity.SecurityHeadersMiddleware())

	// CORS middleware (if enabled) - must be early to handle preflight requests
	corsCfg := config.LoadCORSConfig()
	if corsCfg.Enabled {
		slogLogger.Info("CORS enabled", "allowed_origins", corsCfg.AllowedOrigins, "allowed_methods", corsCfg.AllowedMethods, "allow_credentials", corsCfg.AllowCredentials)
		config.SetupCORS(app, corsCfg)
	} else {
		slogLogger.Info("CORS disabled")
	}

	// Request timeout middleware (30 seconds default)
	app.Use(apptimeout.Middleware(apptimeout.DefaultConfig()))

	// Add OpenTelemetry tracing middleware (must be first to extract trace context)
	app.Use(apptracing.Middleware(cfg.AppName))
	// Add request ID middleware for distributed tracing (works with tracing, preserves trace_id)
	app.Use(applogger.RequestIDMiddleware(slogLogger))
	// Add structured request logging middleware
	app.Use(applogger.RequestLoggerMiddleware(slogLogger))
	// Add Prometheus metrics middleware
	app.Use(appmetrics.Middleware())

	// Request size limit (10MB)
	app.Use(appsecurity.RequestSizeLimitMiddleware(10 * 1024 * 1024))

	// Input sanitization
	app.Use(appsecurity.InputSanitizationMiddleware())

	// CSRF protection (for state-changing requests)
	// Secure cookie should be false in development (HTTP) and true in production (HTTPS)
	secureCookie := env == "production"
	csrfCfg := appsecurity.DefaultCSRFConfig(dbManager.GetRedis(), secureCookie)
	app.Use(appsecurity.CSRFMiddleware(csrfCfg))

	// Rate limiting (100 requests per minute per IP/user)
	app.Use(appsecurity.RateLimitMiddleware(100, time.Minute))

	// Initialize health checker
	healthChecker := health.NewHealthChecker(dbManager.GetPostgres(), dbManager.GetRedis(), slogLogger)

	// Health check endpoints (before API routes, no auth required)
	app.Get("/healthz", healthChecker.Handler())                // Combined health check
	app.Get("/healthz/live", healthChecker.LivenessHandler())   // Liveness probe (Kubernetes)
	app.Get("/healthz/ready", healthChecker.ReadinessHandler()) // Readiness probe (Kubernetes)

	// Prometheus metrics endpoint (before API routes, no auth required)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// API routes group
	api := app.Group("/api/v1")

	// CSRF token endpoint (GET request - middleware will generate token automatically)
	api.Get("/csrf-token", func(c *fiber.Ctx) error {
		// Token is already set in header by CSRF middleware
		// Just return a simple success response
		return c.JSON(fiber.Map{
			"success": true,
			"message": "CSRF token available in X-CSRF-Token header",
		})
	})

	// Load AI service URL for cover letter generation
	aiServiceURL := os.Getenv("AI_SERVICE_URL")
	if aiServiceURL == "" {
		aiServiceURL = "http://ai-service:8000"
	}

	// Initialize JWT manager for token validation (shared secret with auth service)
	authCfg, err := config.LoadAuthConfig()
	if err != nil {
		slogLogger.Warn("Failed to load auth config, JWT validation may not work", "error", err)
	}
	var jwtManager *authPkg.JWTManager
	if authCfg != nil {
		jwtManager = authPkg.NewJWTManager(
			authCfg.JWTSecret,
			cfg.AppName,
			time.Duration(authCfg.JWTExpireHours)*time.Hour,
			time.Duration(authCfg.JWTRefreshExpireHours)*time.Hour,
		)
		// Set Redis client for token blacklist support
		if dbManager.GetRedis() != nil {
			jwtManager.SetRedisClient(dbManager.GetRedis())
		}
	}

	// Setup jobs domain routes
	jobsdomain.SetupRoutes(api, dbManager, jwtManager, aiServiceURL, slogLogger)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		slogLogger.Info("starting jobs service", "addr", addr, "env", env)
		if err := app.Listen(addr); err != nil {
			slogLogger.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	slogLogger.Info("shutting down jobs service gracefully")

	// Give ongoing requests time to complete
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		slogLogger.Error("error during shutdown", "error", err)
	}

	slogLogger.Info("jobs service stopped")
}
