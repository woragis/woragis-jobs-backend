package middleware

import (
	"errors"
	"log"

	authpkg "woragis-jobs-service/pkg/auth"
	apperrors "woragis-jobs-service/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// JWTConfig holds the configuration for JWT middleware
type JWTConfig struct {
	JWTManager *authpkg.JWTManager
}

// JWTMiddleware creates a Fiber JWT authentication middleware
func JWTMiddleware(config JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// Check for X-API-Key header as alternative for internal services
			apiKey := c.Get("X-API-Key")
			if apiKey == "internal-service-key" {
				// Allow internal service access
				return c.Next()
			}
			appErr := HeaderError("Authorization")
			appErr.WithContext("request_id", requestID)
			log.Printf("[JWT Auth] %s | %s %s | Request-ID: %s", 
				appErr.Code, c.Method(), c.Path(), requestID)
			return apperrors.SendError(c, appErr)
		}

		// Extract token from "Bearer <token>"
		token, err := authpkg.ExtractTokenFromHeader(authHeader)
		if err != nil {
			appErr := TokenFormatError(err.Error())
			appErr.WithContext("request_id", requestID)
			log.Printf("[JWT Auth] %s | %s %s | Reason: %v | Request-ID: %s", 
				appErr.Code, c.Method(), c.Path(), err, requestID)
			return apperrors.SendError(c, appErr)
		}

		// Validate token
		claims, err := config.JWTManager.Validate(token)
		if err != nil {
			// Create token preview for logging (first 20 chars + ... or just first 20 if shorter)
			tokenPreview := ""
			if len(token) > 20 {
				tokenPreview = token[:20] + "..."
			} else if len(token) > 0 {
				tokenPreview = token
			}
			
			// Create structured error with code and context
			appErr := ValidationFailedError(err.Error(), tokenPreview, requestID)
			log.Printf("[JWT Auth] %s | %s %s | Error: %v | Token: %s | Request-ID: %s", 
				appErr.Code, c.Method(), c.Path(), err, tokenPreview, requestID)
			
			// Enhance context with more info
			appErr.WithContext("method", c.Method()).
				WithContext("path", c.Path()).
				WithContext("error_details", err.Error())
			
			return apperrors.SendError(c, appErr)
		}

		// Add user information to context
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRole", claims.Role)
		c.Locals("userName", claims.Name)

		return c.Next()
	}
}

// RequireRole creates a middleware that checks if user has required role
func RequireRole(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")
		if userRole == nil {
			appErr := apperrors.New(apperrors.AUTH_UNAUTHORIZED).
				WithContext("component", "role_middleware").
				WithContext("reason", "user_role_not_found").
				WithContext("required_role", requiredRole)
			return apperrors.SendError(c, appErr)
		}

		role, ok := userRole.(string)
		if !ok || role != requiredRole {
			appErr := apperrors.New(apperrors.AUTH_UNAUTHORIZED).
				WithContext("component", "role_middleware").
				WithContext("reason", "insufficient_permissions").
				WithContext("required_role", requiredRole).
				WithContext("user_role", role)
			return apperrors.SendError(c, appErr)
		}

		return c.Next()
	}
}

// RequireAdmin creates a middleware that requires admin role
func RequireAdmin() fiber.Handler {
	return RequireRole("admin")
}

// RequireModerator creates a middleware that requires moderator or admin role
func RequireModerator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")
		if userRole == nil {
			appErr := apperrors.New(apperrors.AUTH_UNAUTHORIZED).
				WithContext("component", "moderator_middleware").
				WithContext("reason", "user_role_not_found")
			return apperrors.SendError(c, appErr)
		}

		role, ok := userRole.(string)
		if !ok || (role != "moderator" && role != "admin") {
			appErr := apperrors.New(apperrors.AUTH_UNAUTHORIZED).
				WithContext("component", "moderator_middleware").
				WithContext("reason", "insufficient_permissions").
				WithContext("required_roles", []string{"moderator", "admin"}).
				WithContext("user_role", role)
			return apperrors.SendError(c, appErr)
		}

		return c.Next()
	}
}

// OptionalJWTMiddleware creates a middleware that optionally validates JWT
// If token is present and valid, user info is added to context
// If token is missing or invalid, request continues without user info
func OptionalJWTMiddleware(config JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			token, err := authpkg.ExtractTokenFromHeader(authHeader)
			if err == nil {
				claims, err := config.JWTManager.Validate(token)
				if err == nil {
					// Add user information to context
					c.Locals("userID", claims.UserID)
					c.Locals("userEmail", claims.Email)
					c.Locals("userRole", claims.Role)
					c.Locals("userName", claims.Name)
				}
			}
		}
		return c.Next()
	}
}

// GetUserIDFromFiberContext extracts user ID from Fiber context
func GetUserIDFromFiberContext(c *fiber.Ctx) (uuid.UUID, error) {
	userID := c.Locals("userID")
	if userID == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID in context")
	}

	return id, nil
}

// GetUserRoleFromFiberContext extracts user role from Fiber context
func GetUserRoleFromFiberContext(c *fiber.Ctx) (string, error) {
	userRole := c.Locals("userRole")
	if userRole == nil {
		return "", errors.New("user role not found")
	}

	role, ok := userRole.(string)
	if !ok {
		return "", errors.New("invalid user role in context")
	}

	return role, nil
}

// GetUserEmailFromFiberContext extracts user email from Fiber context
func GetUserEmailFromFiberContext(c *fiber.Ctx) (string, error) {
	userEmail := c.Locals("userEmail")
	if userEmail == nil {
		return "", errors.New("user email not found")
	}

	email, ok := userEmail.(string)
	if !ok {
		return "", errors.New("invalid user email in context")
	}

	return email, nil
}

// GetUserNameFromFiberContext extracts user name from Fiber context
func GetUserNameFromFiberContext(c *fiber.Ctx) (string, error) {
	userName := c.Locals("userName")
	if userName == nil {
		return "", errors.New("user name not found")
	}

	name, ok := userName.(string)
	if !ok {
		return "", errors.New("invalid user name in context")
	}

	return name, nil
}
