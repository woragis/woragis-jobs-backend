package middleware

import "woragis-jobs-service/pkg/errors"

// Middleware-specific error codes (using AUTH errors for now, could create MW_ range if needed)
const (
	// Header extraction errors
	ErrNoAuthHeader     = errors.AUTH_TOKEN_MISSING
	ErrInvalidFormat    = errors.AUTH_TOKEN_INVALID_FORMAT
	ErrCSRFMissing      = errors.CSRF_TOKEN_MISSING
	ErrCSRFInvalid      = errors.CSRF_TOKEN_INVALID
)

// HeaderError creates an error for missing/invalid auth headers
func HeaderError(headerName string) *errors.AppError {
	return errors.New(ErrNoAuthHeader).
		WithContext("component", "jwt_middleware").
		WithContext("header", headerName)
}

// TokenFormatError creates an error for invalid token format
func TokenFormatError(reason string) *errors.AppError {
	return errors.New(ErrInvalidFormat).
		WithContext("component", "jwt_middleware").
		WithContext("reason", reason)
}

// ValidationFailedError creates a generic validation failed error with full context
func ValidationFailedError(errMsg string, tokenPreview string, requestID string) *errors.AppError {
	code := errors.AUTH_JWT_INVALID_SIGNATURE // Default to signature error
	
	// Map common JWT library error messages to codes
	switch {
	case contains(errMsg, "malformed"):
		code = errors.AUTH_JWT_MALFORMED
	case contains(errMsg, "expired"):
		code = errors.AUTH_JWT_EXPIRED
	case contains(errMsg, "signature"):
		code = errors.AUTH_JWT_INVALID_SIGNATURE
	case contains(errMsg, "claims"):
		code = errors.AUTH_JWT_MISSING_CLAIMS
	}

	return errors.New(code).
		WithContext("component", "jwt_middleware").
		WithContext("raw_error", errMsg).
		WithContext("token_preview", tokenPreview).
		WithContext("request_id", requestID)
}

// Helper function
func contains(s string, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
