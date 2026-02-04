package dailyobjectives

const (
	ErrCodeInvalidPayload     = 12100
	ErrCodeValidation         = 12101
	ErrCodeRepositoryFailure  = 12102
	ErrCodeNotFound           = 12103
)

const (
	ErrObjectiveNotFound           = "dailyobjectives: objective not found"
	ErrInvalidPayload              = "dailyobjectives: invalid payload"
	ErrTargetSumMismatch           = "dailyobjectives: sum of junior + pleno + senior must equal total target"
	ErrNegativeTargets             = "dailyobjectives: targets must be non-negative"
	ErrInvalidTargets              = "dailyobjectives: invalid targets"
	ErrUnableToFetch               = "dailyobjectives: unable to fetch data"
	ErrUnableToCreate              = "dailyobjectives: unable to create data"
	ErrUnableToUpdate              = "dailyobjectives: unable to update data"
	ErrUnableToDelete              = "dailyobjectives: unable to delete data"
)

type DomainError struct {
	Code    int
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(code int, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeNotFound
	}
	return false
}

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code == ErrCodeValidation
	}
	return false
}
