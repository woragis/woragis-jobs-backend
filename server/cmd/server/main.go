package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

// maskValue returns first 4 characters of a value, rest as asterisks (for logging secrets safely)
func maskValue(val string) string {
	if val == "" {
		return "<not set>"
	}
	if len(val) <= 4 {
		return "****"
	}
	return val[:4] + strings.Repeat("*", len(val)-4)
}

// logEnvironmentVariables logs all relevant environment variables at startup
func logEnvironmentVariables(logger *slog.Logger, env string) {
	logger.Info("=== Environment Configuration ===")
	
	// Required variables
	requiredVars := map[string]string{
		"DATABASE_URL": os.Getenv("DATABASE_URL"),
		"REDIS_URL":    os.Getenv("REDIS_URL"),
		"AES_KEY":      os.Getenv("AES_KEY"),
		"HASH_SALT":    os.Getenv("HASH_SALT"),
	}
	
	// JWT Secret (check both variants)
	jwtSecret := os.Getenv("AUTH_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
		requiredVars["JWT_SECRET"] = jwtSecret
	} else {
		requiredVars["AUTH_JWT_SECRET"] = jwtSecret
	}
	
	logger.Info("Required Variables:")
	for key, val := range requiredVars {
		status := "✓"
		if val == "" {
			status = "✗"
		}
		logger.Info("  "+status+" "+key, "value", maskValue(val))
	}
	
	// Application variables (with defaults)
	logger.Info("Application Variables:")
	appVars := map[string]string{
		"APP_NAME":       os.Getenv("APP_NAME"),
		"PORT":           os.Getenv("PORT"),
		"ENV":            env,
		"APP_PUBLIC_URL": os.Getenv("APP_PUBLIC_URL"),
	}
	for key, val := range appVars {
		status := "✓"
		display := val
		if val == "" {
			status = "○"
			display = "<using default>"
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// Database connection pool settings (optional)
	logger.Info("Database Pool Settings (optional):")
	dbPoolVars := map[string]string{
		"DATABASE_MAX_OPEN_CONNS":    os.Getenv("DATABASE_MAX_OPEN_CONNS"),
		"DATABASE_MAX_IDLE_CONNS":    os.Getenv("DATABASE_MAX_IDLE_CONNS"),
		"DATABASE_MAX_IDLE_TIME":     os.Getenv("DATABASE_MAX_IDLE_TIME"),
		"DATABASE_CONN_MAX_LIFETIME": os.Getenv("DATABASE_CONN_MAX_LIFETIME"),
	}
	for key, val := range dbPoolVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			display = val
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// Redis settings (optional)
	logger.Info("Redis Settings (optional):")
	redisVars := map[string]string{
		"REDIS_PASSWORD": os.Getenv("REDIS_PASSWORD"),
		"REDIS_DB":       os.Getenv("REDIS_DB"),
	}
	for key, val := range redisVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			if key == "REDIS_PASSWORD" {
				display = maskValue(val)
			} else {
				display = val
			}
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// JWT settings (optional)
	logger.Info("JWT Settings (optional):")
	jwtVars := map[string]string{
		"JWT_EXPIRE_HOURS":         os.Getenv("JWT_EXPIRE_HOURS"),
		"JWT_REFRESH_EXPIRE_HOURS": os.Getenv("JWT_REFRESH_EXPIRE_HOURS"),
		"BCRYPT_COST":              os.Getenv("BCRYPT_COST"),
	}
	for key, val := range jwtVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			display = val
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// RabbitMQ settings (optional)
	logger.Info("RabbitMQ Settings (optional):")
	rabbitVars := map[string]string{
		"RABBITMQ_URL":      os.Getenv("RABBITMQ_URL"),
		"RABBITMQ_HOST":     os.Getenv("RABBITMQ_HOST"),
		"RABBITMQ_PORT":     os.Getenv("RABBITMQ_PORT"),
		"RABBITMQ_USER":     os.Getenv("RABBITMQ_USER"),
		"RABBITMQ_PASSWORD": os.Getenv("RABBITMQ_PASSWORD"),
		"RABBITMQ_VHOST":    os.Getenv("RABBITMQ_VHOST"),
	}
	for key, val := range rabbitVars {
		status := "○"
		display := "<not set>"
		if val != "" {
			status = "✓"
			if key == "RABBITMQ_PASSWORD" || key == "RABBITMQ_URL" {
				display = maskValue(val)
			} else {
				display = val
			}
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// SMTP/Email settings (optional)
	logger.Info("SMTP/Email Settings (optional):")
	smtpVars := map[string]string{
		"SMTP_HOST":     os.Getenv("SMTP_HOST"),
		"SMTP_PORT":     os.Getenv("SMTP_PORT"),
		"SMTP_USERNAME": os.Getenv("SMTP_USERNAME"),
		"SMTP_PASSWORD": os.Getenv("SMTP_PASSWORD"),
		"EMAIL_FROM":    os.Getenv("EMAIL_FROM"),
		"SMTP_TLS":      os.Getenv("SMTP_TLS"),
	}
	for key, val := range smtpVars {
		status := "○"
		display := "<not set>"
		if val != "" {
			status = "✓"
			if key == "SMTP_PASSWORD" {
				display = maskValue(val)
			} else {
				display = val
			}
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// CORS settings (optional)
	logger.Info("CORS Settings (optional):")
	corsVars := map[string]string{
		"CORS_ENABLED":           os.Getenv("CORS_ENABLED"),
		"CORS_ALLOWED_ORIGINS":   os.Getenv("CORS_ALLOWED_ORIGINS"),
		"CORS_ALLOWED_METHODS":   os.Getenv("CORS_ALLOWED_METHODS"),
		"CORS_ALLOWED_HEADERS":   os.Getenv("CORS_ALLOWED_HEADERS"),
		"CORS_EXPOSED_HEADERS":   os.Getenv("CORS_EXPOSED_HEADERS"),
		"CORS_ALLOW_CREDENTIALS": os.Getenv("CORS_ALLOW_CREDENTIALS"),
		"CORS_MAX_AGE":           os.Getenv("CORS_MAX_AGE"),
	}
	for key, val := range corsVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			display = val
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// AI Service settings (optional)
	logger.Info("AI Service Settings (optional):")
	aiVars := map[string]string{
		"AI_SERVICE_URL": os.Getenv("AI_SERVICE_URL"),
	}
	for key, val := range aiVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			display = val
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	// Observability settings (optional)
	logger.Info("Observability Settings (optional):")
	obsVars := map[string]string{
		"JAEGER_ENDPOINT": os.Getenv("JAEGER_ENDPOINT"),
	}
	for key, val := range obsVars {
		status := "○"
		display := "<using default>"
		if val != "" {
			status = "✓"
			display = val
		}
		logger.Info("  "+status+" "+key, "value", display)
	}
	
	logger.Info("=== End Configuration ===")
	logger.Info("")
}

// validateRequiredEnvVars checks that all required environment variables are set
func validateRequiredEnvVars(logger *slog.Logger, env string) error {
	required := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"AES_KEY",
		"HASH_SALT",
	}

	// JWT_SECRET is required in production
	if env == "production" {
		jwtSecret := os.Getenv("AUTH_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = os.Getenv("JWT_SECRET")
		}
		if jwtSecret == "" {
			required = append(required, "AUTH_JWT_SECRET or JWT_SECRET")
		}
	}

	var missing []string
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %v", missing)
	}

	logger.Info("environment variables validated successfully")
	return nil
}

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

	// Log all environment variables for visibility
	logEnvironmentVariables(slogLogger, env)

	// Validate required environment variables before proceeding
	if err := validateRequiredEnvVars(slogLogger, env); err != nil {
		slogLogger.Error("missing required environment variables", "error", err)
		os.Exit(1)
	}

	// Initialize OpenTelemetry tracing
	slogLogger.Info("initializing tracing...")
	tracingShutdown, err := apptracing.Init(apptracing.Config{
		ServiceName:    cfg.AppName,
		ServiceVersion: "1.4.1",
		Environment:    env,
		JaegerEndpoint: os.Getenv("JAEGER_ENDPOINT"), // Defaults to http://jaeger:4318
	})
	if err != nil {
		slogLogger.Warn("failed to initialize tracing", "error", err)
	} else {
		slogLogger.Info("tracing initialized successfully", "service", cfg.AppName)
		defer func() {
			if tracingShutdown != nil {
				tracingShutdown()
			}
		}()
	}

	// Load database and Redis configs
	slogLogger.Info("loading database and redis configurations...")
	dbCfg := config.LoadDatabaseConfig()
	redisCfg := config.LoadRedisConfig()

	// Initialize database manager
	slogLogger.Info("initializing database manager...")
	dbManager, err := database.NewFromConfig(dbCfg, redisCfg)
	if err != nil {
		slogLogger.Error("failed to initialize database manager", "error", err)
		os.Exit(1)
	}
	defer dbManager.Close()

	// Perform initial health check
	slogLogger.Info("performing initial health check...")
	if err := dbManager.HealthCheck(); err != nil {
		slogLogger.Warn("Database health check failed", "error", err)
	} else {
		slogLogger.Info("All database connections are healthy")
	}

	// Run migrations
	slogLogger.Info("running database migrations...")
	if err := jobsdomain.MigrateJobsTables(dbManager.GetPostgres()); err != nil {
		slogLogger.Error("failed to run jobs migrations", "error", err)
		os.Exit(1)
	}

	// Create Fiber app
	slogLogger.Info("creating fiber app...")
	app := config.CreateFiberApp(cfg)
	slogLogger.Info("fiber app created successfully")

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
	slogLogger.Info("setting up routes...")
	jobsdomain.SetupRoutes(api, dbManager, jwtManager, aiServiceURL, slogLogger)
	slogLogger.Info("routes configured successfully")

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		slogLogger.Info("attempting to start server", "addr", addr, "env", env)
		if err := app.Listen(addr); err != nil {
			slogLogger.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Give the server a moment to start and then log ready message
	time.Sleep(100 * time.Millisecond)
	slogLogger.Info("✓ SERVER READY - Jobs service is listening and accepting connections", "port", cfg.Port, "env", env)

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
