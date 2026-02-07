package jobapplications

import (
	"errors"
	"fmt"
	"strings"
)

// Error codes organized by domain with clear prefixes for trackability
const (
	// Job Application errors (APP prefix)
	ErrCodeInvalidPayload              = "APP001"
	ErrCodeInvalidStatus               = "APP002"
	ErrCodeNotFound                    = "APP003"
	ErrCodeAccessDenied                = "APP004"
	ErrCodeDuplicateApplication        = "APP005"
	ErrCodeJobApplicationFailed        = "APP006"
	ErrCodeInvalidResumeID             = "APP007"
	ErrCodeInsufficientData            = "APP008"

	// Job Application operations (APPOP prefix)
	ErrCodeApplicationCreateFailed     = "APPOP001"
	ErrCodeApplicationUpdateFailed     = "APPOP002"
	ErrCodeApplicationDeleteFailed     = "APPOP003"
	ErrCodeApplicationFetchFailed      = "APPOP004"
	ErrCodeStatusUpdateFailed          = "APPOP005"

	// External service errors (EXT prefix)
	ErrCodeAIServiceUnavailable        = "EXT001"
	ErrCodeAIServiceFailed             = "EXT002"
	ErrCodeJobQueueUnavailable         = "EXT003"
	ErrCodePlaywrightUnavailable       = "EXT004"

	// Database errors (DB prefix)
	ErrCodeDBQueryFailed               = "DB001"
	ErrCodeDBWriteFailed               = "DB002"
	ErrCodeDBConnectionFail            = "DB003"
	ErrCodeDatabaseConstraint          = "DB004"
	ErrCodeDatabaseValueTooLong        = "DB005"
	ErrCodeDatabaseUniqueViolation     = "DB006"
	ErrCodeDatabaseForeignKeyViolation = "DB007"

	// Validation errors (VAL prefix)
	ErrCodeValidationFailed            = "VAL001"
	ErrCodeMissingField                = "VAL002"
	ErrCodeInvalidField                = "VAL003"
	ErrCodeInvalidUUID                 = "VAL004"

	// Backward compat - old numeric codes
	ErrCodeOldInvalidPayload       = 10000
	ErrCodeOldInvalidStatus        = 10001
	ErrCodeOldRepositoryFailure    = 10002
	ErrCodeOldNotFound             = 10003
	ErrCodeOldJobQueueFailure      = 10004
	ErrCodeOldAIServiceFailure     = 10005
	ErrCodeOldPlaywrightFailure    = 10006
	ErrCodeOldAccessDenied         = 10007
	ErrCodeOldDatabaseConstraint   = 10008
	ErrCodeOldDatabaseValueTooLong = 10009
	ErrCodeOldDatabaseUniqueViolation = 10010
	ErrCodeOldDatabaseForeignKeyViolation = 10011
	ErrCodeOldDatabaseConnection   = 10012
)

// Old string constants for backward compatibility
const (
	ErrNilApplication                = "jobapplications: application entity is nil"
	ErrEmptyApplicationID            = "jobapplications: application id cannot be empty"
	ErrEmptyUserID                   = "jobapplications: user id cannot be empty"
	ErrEmptyCompanyName              = "jobapplications: company name cannot be empty"
	ErrEmptyJobTitle                 = "jobapplications: job title cannot be empty"
	ErrEmptyJobURL                   = "jobapplications: job url cannot be empty"
	ErrEmptyWebsite                  = "jobapplications: website cannot be empty"
	ErrApplicationNotFound           = "jobapplications: application not found"
	ErrUnsupportedStatus             = "jobapplications: unsupported status"
	ErrUnableToPersist               = "jobapplications: unable to persist data"
	ErrUnableToFetch                 = "jobapplications: unable to fetch data"
	ErrUnableToUpdate                = "jobapplications: unable to update data"
	ErrJobQueueUnavailable           = "jobapplications: job queue unavailable"
	ErrAIServiceUnavailable          = "jobapplications: AI service unavailable"
	ErrPlaywrightUnavailable         = "jobapplications: Playwright unavailable"
	ErrJobApplicationFailed          = "jobapplications: job application failed"
	ErrDatabaseConstraintViolation   = "jobapplications: database constraint violation"
	ErrValueTooLong                  = "jobapplications: one or more field values exceed the maximum allowed length"
	ErrDatabaseUniqueViolation       = "jobapplications: a record with this information already exists"
	ErrDatabaseForeignKeyViolation   = "jobapplications: referenced record does not exist"
	ErrDatabaseConnectionFailure     = "jobapplications: database connection error"
)

// ErrorMessages maps error codes to user-friendly messages
var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:              "Invalid request payload",
	ErrCodeInvalidStatus:               "Invalid application status",
	ErrCodeNotFound:                    "Application not found",
	ErrCodeAccessDenied:                "Access denied to this application",
	ErrCodeDuplicateApplication:        "A record with this information already exists",
	ErrCodeJobApplicationFailed:        "Job application operation failed",
	ErrCodeInvalidResumeID:             "Invalid or missing resume ID",
	ErrCodeInsufficientData:            "Insufficient data to process request",

	ErrCodeApplicationCreateFailed:     "Failed to create application",
	ErrCodeApplicationUpdateFailed:     "Failed to update application",
	ErrCodeApplicationDeleteFailed:     "Failed to delete application",
	ErrCodeApplicationFetchFailed:      "Failed to fetch application",
	ErrCodeStatusUpdateFailed:          "Failed to update application status",

	ErrCodeAIServiceUnavailable:        "AI service is unavailable",
	ErrCodeAIServiceFailed:             "AI service request failed",
	ErrCodeJobQueueUnavailable:         "Job queue is unavailable",
	ErrCodePlaywrightUnavailable:       "Playwright service is unavailable",

	ErrCodeDBQueryFailed:               "Database query failed",
	ErrCodeDBWriteFailed:               "Database write failed",
	ErrCodeDBConnectionFail:            "Database connection failed",
	ErrCodeDatabaseConstraint:          "Database constraint violation",
	ErrCodeDatabaseValueTooLong:        "One or more field values exceed maximum length",
	ErrCodeDatabaseUniqueViolation:     "A record with this information already exists",
	ErrCodeDatabaseForeignKeyViolation: "Referenced record does not exist",

	ErrCodeValidationFailed:            "Validation failed",
	ErrCodeMissingField:                "Required field missing",
	ErrCodeInvalidField:                "Invalid field value",
	ErrCodeInvalidUUID:                 "Invalid UUID format",
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
// For backward compatibility, this also accepts old string error messages
func NewDomainError(code string, err error, context ...map[string]interface{}) *DomainError {
	msg, ok := ErrorMessages[code]
	if !ok {
		// Fallback: if code looks like an old error string constant, use it
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
	case ErrCodeInvalidPayload, ErrCodeMissingField, ErrCodeInvalidField, ErrCodeInvalidUUID, ErrCodeValidationFailed, ErrCodeInvalidStatus:
		return 400
	case ErrCodeNotFound, ErrCodeApplicationFetchFailed:
		return 404
	case ErrCodeAccessDenied:
		return 403
	case ErrCodeDuplicateApplication, ErrCodeDatabaseUniqueViolation:
		return 409
	case ErrCodeAIServiceUnavailable, ErrCodeAIServiceFailed, ErrCodeJobQueueUnavailable, ErrCodePlaywrightUnavailable:
		return 502
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

// handleDatabaseError converts database errors to domain errors.
// It checks for PostgreSQL error codes (SQLSTATE) in the error message.
func handleDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already a DomainError, return it
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr
	}

	errStr := err.Error()

	// Check for PostgreSQL SQLSTATE codes
	// SQLSTATE 22001: string data right truncated / value too long
	if strings.Contains(errStr, "SQLSTATE 22001") || strings.Contains(errStr, "value too long") {
		return NewDomainError(ErrCodeDatabaseValueTooLong, err, map[string]interface{}{
			"sqlstate": "22001",
		})
	}

	// SQLSTATE 23505: unique violation
	if strings.Contains(errStr, "SQLSTATE 23505") || strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
		return NewDomainError(ErrCodeDatabaseUniqueViolation, err, map[string]interface{}{
			"sqlstate": "23505",
		})
	}

	// SQLSTATE 23503: foreign key violation
	if strings.Contains(errStr, "SQLSTATE 23503") || strings.Contains(errStr, "foreign key constraint") {
		return NewDomainError(ErrCodeDatabaseForeignKeyViolation, err, map[string]interface{}{
			"sqlstate": "23503",
		})
	}

	// SQLSTATE 23514: check constraint violation
	// SQLSTATE 23502: not null violation
	// SQLSTATE 23XXX: other constraint violations
	if strings.Contains(errStr, "SQLSTATE 23") || strings.Contains(errStr, "constraint") {
		return NewDomainError(ErrCodeDatabaseConstraint, err, map[string]interface{}{
			"sqlstate": "23XXX",
		})
	}

	// Connection errors
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "dial") || strings.Contains(errStr, "network") {
		return NewDomainError(ErrCodeDBConnectionFail, err, map[string]interface{}{
			"type": "connection",
		})
	}

	// For any other database error, return a generic DB write failure
	return NewDomainError(ErrCodeDBWriteFailed, err, map[string]interface{}{
		"raw_error": errStr,
	})
}

// Deprecated: Use NewDomainError with proper error type instead. This is for backward compatibility.
// Wraps a string message into a DomainError
func NewDomainErrorFromString(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
		err:     fmt.Errorf(message),
	}
}

