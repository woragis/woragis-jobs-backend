package resumes

import (
	"fmt"
)

// Old string constants for backward compatibility with validation code
const (
	ErrNilResume        = "resumes: resume cannot be nil"
	ErrEmptyResumeID    = "resumes: resume ID cannot be empty"
	ErrEmptyUserID      = "resumes: user ID cannot be empty"
	ErrEmptyResumeTitle = "resumes: resume title cannot be empty"
	ErrEmptyFilePath    = "resumes: file path cannot be empty"
	ErrEmptyFileName    = "resumes: file name cannot be empty"
	ErrInvalidFileSize  = "resumes: file size cannot be negative"
	ErrResumeNotFound   = "resumes: resume not found"
	ErrFileNotFound     = "resumes: resume file not found"
	ErrFileReadError    = "resumes: error reading resume file"
	ErrNoMainResume     = "resumes: no main resume found"
)

// Error codes organized by domain with clear prefixes for trackability
const (
	// Resume generation errors (RES prefix)
	ErrCodeGenerationFailed        = "RES001"
	ErrCodeQueueFailed             = "RES002"
	ErrCodeJobNotFound             = "RES003"
	ErrCodeJobAlreadyCompleted     = "RES004"
	ErrCodeRetryFailed             = "RES005"
	ErrCodePublisherNotInitialized = "RES006"
	ErrCodePublisherFailed         = "RES007"
	ErrCodeResumeNotFound          = "RES008"
	ErrCodeInvalidJobStatus        = "RES009"
	ErrCodeCancelFailed            = "RES010"

	// Resume CRUD errors (RESRC prefix)
	ErrCodeResumeCreateFailed      = "RESRC001"
	ErrCodeResumeUpdateFailed      = "RESRC002"
	ErrCodeResumeDeleteFailed      = "RESRC003"
	ErrCodeResumeValidationFailed  = "RESRC004"

	// File errors (FILE prefix)
	ErrCodeFileNotFound   = "FILE001"
	ErrCodeFileReadError  = "FILE002"
	ErrCodeInvalidFileSize = "FILE003"

	// Database errors (DB prefix)
	ErrCodeDBQueryFailed    = "DB001"
	ErrCodeDBWriteFailed    = "DB002"
	ErrCodeDBConnectionFail = "DB003"

	// HTTP/Network errors (HTTP prefix)
	ErrCodeHTTPRequestFailed  = "HTTP001"
	ErrCodeHTTPResponseFailed = "HTTP002"
	ErrCodeNetworkTimeout     = "HTTP003"

	// Validation errors (VAL prefix)
	ErrCodeInvalidPayload = "VAL001"
	ErrCodeMissingField   = "VAL002"
	ErrCodeInvalidUUID    = "VAL003"

	// Backwards compat - old error codes still used in some places
	ErrCodeInvalidName = "INVALID_NAME"
	ErrCodeNotFound    = "NOT_FOUND"
)

// ErrorMessages maps error codes to user-friendly messages
var ErrorMessages = map[string]string{
	ErrCodeGenerationFailed:        "Resume generation failed",
	ErrCodeQueueFailed:             "Failed to queue resume generation job",
	ErrCodeJobNotFound:             "Resume generation job not found",
	ErrCodeJobAlreadyCompleted:     "Job has already been completed",
	ErrCodeRetryFailed:             "Failed to retry resume generation",
	ErrCodePublisherNotInitialized: "Resume publisher not configured",
	ErrCodePublisherFailed:         "Failed to submit job to resume generator",
	ErrCodeResumeNotFound:          "Resume not found",
	ErrCodeInvalidJobStatus:        "Invalid job status for this operation",
	ErrCodeCancelFailed:            "Failed to cancel resume generation job",

	ErrCodeResumeCreateFailed:     "Failed to create resume",
	ErrCodeResumeUpdateFailed:     "Failed to update resume",
	ErrCodeResumeDeleteFailed:     "Failed to delete resume",
	ErrCodeResumeValidationFailed: "Resume validation failed",

	ErrCodeFileNotFound:     "Resume file not found",
	ErrCodeFileReadError:    "Error reading resume file",
	ErrCodeInvalidFileSize:  "Invalid file size",

	ErrCodeDBQueryFailed:    "Database query failed",
	ErrCodeDBWriteFailed:    "Database write failed",
	ErrCodeDBConnectionFail: "Database connection failed",

	ErrCodeHTTPRequestFailed:  "HTTP request failed",
	ErrCodeHTTPResponseFailed: "HTTP response error",
	ErrCodeNetworkTimeout:     "Network request timeout",

	ErrCodeInvalidPayload: "Invalid request payload",
	ErrCodeMissingField:   "Required field missing",
	ErrCodeInvalidUUID:    "Invalid UUID format",

	ErrCodeInvalidName: "Invalid name",
	ErrCodeNotFound:    "Resource not found",
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
	case ErrCodeInvalidPayload, ErrCodeMissingField, ErrCodeInvalidUUID, ErrCodeResumeValidationFailed:
		return 400
	case ErrCodeJobNotFound, ErrCodeResumeNotFound, ErrCodeFileNotFound, ErrCodeNotFound:
		return 404
	case ErrCodeJobAlreadyCompleted, ErrCodeInvalidJobStatus:
		return 409
	case ErrCodeNetworkTimeout, ErrCodeHTTPResponseFailed, ErrCodePublisherFailed:
		return 502
	default:
		return 500
	}
}

// GetErrorCode returns the error code if this is a DomainError, otherwise returns a generic code
func GetErrorCode(err error) string {
	if de, ok := err.(*DomainError); ok {
		return de.Code
	}
	return "UNKNOWN"
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

