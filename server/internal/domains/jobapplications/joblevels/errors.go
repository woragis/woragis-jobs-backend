package joblevels

import (
	"errors"
	"fmt"
	"strings"
)

// Error codes organized by domain with clear prefixes for trackability
const (
	// Job Level errors (JOB prefix)
	ErrCodeInvalidPayload           = "JOB001"
	ErrCodeInvalidSeniority         = "JOB002"
	ErrCodeInvalidIntensity         = "JOB003"
	ErrCodeNotFound                 = "JOB004"
	ErrCodeDatabaseError            = "JOB005"

	// Backward compat - old numeric codes
	ErrCodeOldInvalidPayload    = 11100
	ErrCodeOldInvalidSeniority  = 11101
	ErrCodeOldInvalidIntensity  = 11102
	ErrCodeOldRepositoryFailure = 11103
	ErrCodeOldNotFound          = 11104
)

// Old string constants for backward compatibility
const (
	ErrJobLevelNotFound = "joblevels: job level not found"
	ErrInvalidSeniority = "joblevels: invalid seniority level"
	ErrInvalidIntensity = "joblevels: invalid intensity level"
	ErrInvalidName      = "joblevels: invalid job level name"
	ErrUnableToFetch    = "joblevels: unable to fetch data"
	ErrUnableToCreate   = "joblevels: unable to create data"
)

// ErrorMessages maps error codes to user-friendly messages
var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:   "Invalid request payload",
	ErrCodeInvalidSeniority: "Invalid seniority level",
	ErrCodeInvalidIntensity: "Invalid intensity level",
	ErrCodeNotFound:         "Job level not found",
	ErrCodeDatabaseError:    "Database operation failed",
}

// DomainError represents a domain-specific error with code, message, and context
type DomainError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
	err     error                  // Internal error for logging
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("%s: %s | %v", e.Code, e.Message, e.Context)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.err
}

// NewDomainError creates a new domain error with code and optional context
func NewDomainError(code string, err error, context ...map[string]interface{}) *DomainError {
	msg, ok := ErrorMessages[code]
	if !ok {
		if code != "" {
			msg = code
		} else {
			msg = "Unknown error"
		}
	}

	ctx := make(map[string]interface{})
	if len(context) > 0 {
		ctx = context[0]
	}

	return &DomainError{
		Code:    code,
		Message: msg,
		Context: ctx,
		err:     err,
	}
}

// GetHTTPStatus returns appropriate HTTP status code for the error
func (e *DomainError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidPayload, ErrCodeInvalidSeniority, ErrCodeInvalidIntensity:
		return 400
	case ErrCodeNotFound:
		return 404
	default:
		return 500
	}
}

// AsDomainError type-asserts an error to DomainError
func AsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Code == ErrCodeNotFound
	}
	return false
}

// handleDatabaseError converts database errors to domain errors
func handleDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already a DomainError, return it
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr
	}

	errStr := err.Error()

	// Connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "dial") || strings.Contains(errStr, "network") {
		return NewDomainError(ErrCodeDatabaseError, err, map[string]interface{}{
			"type": "connection",
		})
	}

	// For any other database error, return a generic database error
	return NewDomainError(ErrCodeDatabaseError, err, map[string]interface{}{
		"raw_error": errStr,
	})
}
