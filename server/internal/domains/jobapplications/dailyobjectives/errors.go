package dailyobjectives

import (
	"errors"
	"fmt"
)

const (
	ErrCodeInvalidPayload    = "OBJ001"
	ErrCodeValidation        = "OBJ002"
	ErrCodeRepositoryFailure = "OBJ003"
	ErrCodeNotFound          = "OBJ004"
)

const (
	ErrObjectiveNotFound = "dailyobjectives: objective not found"
	ErrInvalidPayload    = "dailyobjectives: invalid payload"
	ErrTargetSumMismatch = "dailyobjectives: sum of junior + pleno + senior must equal total target"
	ErrNegativeTargets   = "dailyobjectives: targets must be non-negative"
	ErrInvalidTargets    = "dailyobjectives: invalid targets"
	ErrUnableToFetch     = "dailyobjectives: unable to fetch data"
	ErrUnableToCreate    = "dailyobjectives: unable to create data"
	ErrUnableToUpdate    = "dailyobjectives: unable to update data"
	ErrUnableToDelete    = "dailyobjectives: unable to delete data"
)

var ErrorMessages = map[string]string{
	ErrCodeInvalidPayload:    "Invalid request payload",
	ErrCodeValidation:        "Validation failed",
	ErrCodeRepositoryFailure: "Repository operation failed",
	ErrCodeNotFound:          "Objective not found",
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
	case ErrCodeInvalidPayload, ErrCodeValidation:
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

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Code == ErrCodeNotFound
	}
	return false
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := AsDomainError(err); ok {
		return domainErr.Code == ErrCodeValidation
	}
	return false
}
