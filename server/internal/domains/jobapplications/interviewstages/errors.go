package interviewstages

import (
	"errors"
	"fmt"
)

const (
	ErrCodeInvalidPayload    = "STG001"
	ErrCodeInvalidStageType  = "STG002"
	ErrCodeInvalidOutcome    = "STG003"
	ErrCodeRepositoryFailure = "STG004"
	ErrCodeNotFound          = "STG005"
)

const (
	ErrNilStage                 = "interviewstages: stage entity is nil"
	ErrEmptyStageID            = "interviewstages: stage id cannot be empty"
	ErrEmptyJobApplicationID   = "interviewstages: job application id cannot be empty"
	ErrStageNotFound           = "interviewstages: stage not found"
	ErrUnsupportedStageType    = "interviewstages: unsupported stage type"
	ErrUnsupportedOutcome      = "interviewstages: unsupported outcome"
	ErrUnableToPersist         = "interviewstages: unable to persist data"
	ErrUnableToFetch           = "interviewstages: unable to fetch data"
	ErrUnableToUpdate          = "interviewstages: unable to update data"
)

var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:    "Invalid request payload",
	ErrCodeInvalidStageType:  "Invalid stage type",
	ErrCodeInvalidOutcome:    "Invalid outcome",
	ErrCodeRepositoryFailure: "Repository operation failed",
	ErrCodeNotFound:          "Stage not found",
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
	case ErrCodeInvalidPayload, ErrCodeInvalidStageType, ErrCodeInvalidOutcome:
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

