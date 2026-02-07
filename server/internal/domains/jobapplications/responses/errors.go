package responses

import (
	"errors"
	"fmt"
)

const (
	ErrCodeInvalidPayload        = "RSP001"
	ErrCodeInvalidResponseType   = "RSP002"
	ErrCodeRepositoryFailure     = "RSP003"
	ErrCodeNotFound              = "RSP004"
)

const (
	ErrNilResponse                = "responses: response entity is nil"
	ErrEmptyResponseID            = "responses: response id cannot be empty"
	ErrEmptyJobApplicationID      = "responses: job application id cannot be empty"
	ErrResponseNotFound           = "responses: response not found"
	ErrUnsupportedResponseType    = "responses: unsupported response type"
	ErrUnableToPersist            = "responses: unable to persist data"
	ErrUnableToFetch              = "responses: unable to fetch data"
	ErrUnableToUpdate             = "responses: unable to update data"
)

var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:        "Invalid request payload",
	ErrCodeInvalidResponseType:   "Invalid response type",
	ErrCodeRepositoryFailure:     "Repository operation failed",
	ErrCodeNotFound:              "Response not found",
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
	case ErrCodeInvalidPayload, ErrCodeInvalidResponseType:
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

