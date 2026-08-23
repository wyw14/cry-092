package shared

import "errors"

var (
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrInvalidState     = errors.New("invalid state transition")
	ErrVersionConflict  = errors.New("optimistic version conflict")
	ErrIdempotencyReuse = errors.New("idempotency key reused with different input")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type DomainError struct {
	Code    string
	Message string
	Fields  []FieldError
	Cause   error
}

func (e *DomainError) Error() string { return e.Message }
func (e *DomainError) Unwrap() error { return e.Cause }

func NewError(code, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}
