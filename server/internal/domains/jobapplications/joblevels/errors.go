package joblevels

const (
	ErrCodeInvalidPayload    = 11100
	ErrCodeInvalidSeniority  = 11101
	ErrCodeInvalidIntensity  = 11102
	ErrCodeRepositoryFailure = 11103
	ErrCodeNotFound          = 11104
)

const (
	ErrJobLevelNotFound = "joblevels: job level not found"
	ErrInvalidSeniority = "joblevels: invalid seniority level"
	ErrInvalidIntensity = "joblevels: invalid intensity level"
	ErrInvalidName      = "joblevels: invalid job level name"
	ErrUnableToFetch    = "joblevels: unable to fetch data"
	ErrUnableToCreate   = "joblevels: unable to create data"
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
