package jobwebsites

import (
	"errors"
	"fmt"
)

const (
	ErrCodeInvalidPayload    = "WEB001"
	ErrCodeRepositoryFailure = "WEB002"
	ErrCodeNotFound          = "WEB003"
)

const (
	ErrNilWebsite          = "jobwebsites: website entity is nil"
	ErrEmptyWebsiteID      = "jobwebsites: website id cannot be empty"
	ErrEmptyWebsiteName    = "jobwebsites: website name cannot be empty"
	ErrEmptyDisplayName    = "jobwebsites: display name cannot be empty"
	ErrWebsiteNotFound     = "jobwebsites: website not found"
	ErrInvalidDailyLimit   = "jobwebsites: daily limit cannot be negative"
	ErrInvalidCurrentCount = "jobwebsites: current count cannot be negative"
	ErrUnableToPersist     = "jobwebsites: unable to persist data"
	ErrUnableToFetch       = "jobwebsites: unable to fetch data"
	ErrUnableToUpdate      = "jobwebsites: unable to update data"
)

var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:    "Invalid request payload",
	ErrCodeRepositoryFailure: "Repository operation failed",
	ErrCodeNotFound:          "Website not found",
}

type DomainError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
	err     error                  // Internal error for logging
}

func (e *DomainError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("%s: %s | %v", e.Code, e.Message, e.Context)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.err
}

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

func (e *DomainError) GetHTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidPayload:
		return 400
	case ErrCodeNotFound:
		return 404
	default:
		return 500
	}
}

func AsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}

