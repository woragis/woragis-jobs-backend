package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCSRFTokenMissing = errors.New("CSRF token is missing")
	ErrCSRFTokenInvalid = errors.New("CSRF token is invalid")
	ErrCSRFTokenExpired = errors.New("CSRF token has expired")
)

// CSRFConfig holds configuration for CSRF protection
type CSRFConfig struct {
	// RedisClient is the Redis client for storing CSRF tokens
	RedisClient *redis.Client
	// TokenLength is the length of the CSRF token (default: 32 bytes)
	TokenLength int
	// TokenTTL is the time-to-live for CSRF tokens (default: 1 hour)
	TokenTTL time.Duration
	// CookieName is the name of the cookie to store CSRF token (default: "csrf_token")
	CookieName string
	// HeaderName is the name of the header to read CSRF token (default: "X-CSRF-Token")
	HeaderName string
	// SecureCookie determines if the cookie should only be sent over HTTPS (default: true in production)
	SecureCookie bool
	// ExemptRoutes is a list of routes that don't require CSRF protection
	ExemptRoutes []string
	// ExemptMethods is a list of HTTP methods that don't require CSRF protection
	ExemptMethods []string
}

// DefaultCSRFConfig returns default CSRF configuration
// secureCookie should be false in development (HTTP) and true in production (HTTPS)
func DefaultCSRFConfig(redisClient *redis.Client, secureCookie bool) CSRFConfig {
	return CSRFConfig{
		RedisClient:  redisClient,
		TokenLength:  32,
		TokenTTL:     1 * time.Hour,
		CookieName:   "csrf_token",
		HeaderName:   "X-CSRF-Token",
		SecureCookie: secureCookie,
		ExemptRoutes: []string{"/healthz", "/metrics", "/api/v1/auth/login", "/api/v1/auth/register"},
		ExemptMethods: []string{"GET", "HEAD", "OPTIONS"},
	}
}

// generateToken generates a cryptographically secure random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CSRFMiddleware creates a CSRF protection middleware
func CSRFMiddleware(config CSRFConfig) fiber.Handler {
	// Build exempt methods map for quick lookup
	exemptMethods := make(map[string]bool)
	for _, method := range config.ExemptMethods {
		exemptMethods[method] = true
	}

	// Build exempt routes map for quick lookup
	exemptRoutes := make(map[string]bool)
	for _, route := range config.ExemptRoutes {
		exemptRoutes[route] = true
	}

	return func(c *fiber.Ctx) error {
		// For GET requests, always generate and return a new CSRF token
		// This happens before exemption checks so tokens are always available
		if c.Method() == "GET" {
			token, err := generateToken(config.TokenLength)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to generate CSRF token",
				})
			}

			// Store token in Redis
			if config.RedisClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// Use session ID or IP address as key
				sessionKey := fmt.Sprintf("csrf:%s", c.IP())
				if err := config.RedisClient.Set(ctx, sessionKey, token, config.TokenTTL).Err(); err != nil {
					// Log error but continue (graceful degradation)
					c.Locals("csrf_error", "Failed to store CSRF token")
				}
			}

			// Set token in cookie
			c.Cookie(&fiber.Cookie{
				Name:     config.CookieName,
				Value:    token,
				HTTPOnly: false, // Must be readable by JavaScript for API clients
				Secure:   config.SecureCookie, // Configurable based on environment
				SameSite: "Strict",
				MaxAge:   int(config.TokenTTL.Seconds()),
			})

			// Also set in response header for API clients
			c.Set(config.HeaderName, token)

			return c.Next()
		}

		// For state-changing requests (POST, PUT, PATCH, DELETE), validate CSRF token
		// Check if method is exempt from validation (e.g., OPTIONS for CORS preflight)
		if exemptMethods[c.Method()] {
			return c.Next()
		}

		// Check if route is exempt from validation
		if exemptRoutes[c.Path()] {
			return c.Next()
		}

		// Extract token from header or cookie
		token := c.Get(config.HeaderName)
		if token == "" {
			// Try to get from cookie
			token = c.Cookies(config.CookieName)
		}

		if token == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": ErrCSRFTokenMissing.Error(),
			})
		}

		// Validate token in Redis
		if config.RedisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			sessionKey := fmt.Sprintf("csrf:%s", c.IP())
			storedToken, err := config.RedisClient.Get(ctx, sessionKey).Result()
			if err == redis.Nil {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": ErrCSRFTokenExpired.Error(),
				})
			}
			if err != nil {
				// Redis error - log but allow request (graceful degradation)
				c.Locals("csrf_warning", "CSRF token validation skipped due to Redis error")
				return c.Next()
			}

			if storedToken != token {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": ErrCSRFTokenInvalid.Error(),
				})
			}

			// Token is valid, extend TTL
			config.RedisClient.Expire(ctx, sessionKey, config.TokenTTL)
		}

		return c.Next()
	}
}

