package auth

import "woragis-jobs-service/pkg/errors"

// Auth-specific error codes
const (
	// JWT validation errors
	ErrJWTMalformed       = errors.AUTH_JWT_MALFORMED
	ErrJWTExpired         = errors.AUTH_JWT_EXPIRED
	ErrJWTInvalidSig      = errors.AUTH_JWT_INVALID_SIGNATURE
	ErrJWTMissingClaims   = errors.AUTH_JWT_MISSING_CLAIMS
	ErrTokenMissing       = errors.AUTH_TOKEN_MISSING
	ErrTokenInvalidFormat = errors.AUTH_TOKEN_INVALID_FORMAT
	ErrUnauthorized       = errors.AUTH_UNAUTHORIZED
	ErrServiceUnavail     = errors.AUTH_SERVICE_UNAVAILABLE
)

// AuthError creates an auth-specific AppError
// It automatically maps JWT parse errors to correct error codes
func AuthError(code string, ctx map[string]interface{}) *errors.AppError {
	return errors.New(code).WithContext("component", "auth")
}

// JWTValidationError creates a JWT validation error with token details
func JWTValidationError(reason string, tokenPreview string) *errors.AppError {
	errCode := ErrJWTMalformed
	switch reason {
	case "expired":
		errCode = ErrJWTExpired
	case "invalid_signature":
		errCode = ErrJWTInvalidSig
	case "missing_claims":
		errCode = ErrJWTMissingClaims
	}

	return errors.New(errCode).
		WithContext("reason", reason).
		WithContext("token_preview", tokenPreview).
		WithContext("component", "jwt_validator")
}
