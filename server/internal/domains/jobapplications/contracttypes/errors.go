package contracttypes

const (
	ErrCodeInvalidPayload     = 11200
	ErrCodeInvalidName        = 11201
	ErrCodeRepositoryFailure  = 11202
	ErrCodeNotFound           = 11203
)

const (
	ErrContractTypeNotFound = "contracttypes: contract type not found"
	ErrInvalidName          = "contracttypes: invalid contract type name"
	ErrUnableToFetch        = "contracttypes: unable to fetch data"
	ErrUnableToCreate       = "contracttypes: unable to create data"
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
