package errdefs

import (
	"fmt"
	"net/http"
)

type ErrorType int

const (
	ErrTypeInvalidInput ErrorType = iota
	ErrTypeNotFound
	ErrTypeConflict
	ErrTypeUnauthorized
	ErrTypeProvider
	ErrTypeUnsupported
	ErrTypeInternal
)

// AppErrorBase is where app-specific ErrorType values start; register them
// with Register during init, before any concurrent use.
const AppErrorBase ErrorType = 1000

type registration struct {
	name   string
	status int
}

var registry = map[ErrorType]registration{
	ErrTypeInvalidInput: {"ErrInvalidInput", http.StatusBadRequest},
	ErrTypeNotFound:     {"ErrNotFound", http.StatusNotFound},
	ErrTypeConflict:     {"ErrConflict", http.StatusConflict},
	ErrTypeUnauthorized: {"ErrUnauthorized", http.StatusUnauthorized},
	ErrTypeProvider:     {"ErrProvider", http.StatusInternalServerError},
	ErrTypeUnsupported:  {"ErrUnsupported", http.StatusNotImplemented},
	ErrTypeInternal:     {"ErrInternal", http.StatusInternalServerError},
}

func Register(t ErrorType, name string, status int) {
	registry[t] = registration{name: name, status: status}
}

func (e ErrorType) String() string {
	r, ok := registry[e]
	if !ok {
		return "ErrUnknown"
	}
	return r.name
}

type CustomError struct {
	Type    ErrorType
	Message string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

func (e *CustomError) GetStatus() int {
	r, ok := registry[e.Type]
	if !ok {
		return http.StatusInternalServerError
	}
	return r.status
}

func (e *CustomError) Is(target error) bool {
	t, ok := target.(*CustomError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

func NewCustomError(t ErrorType, message string) *CustomError {
	return &CustomError{Type: t, Message: message}
}

var (
	ErrInvalidInput = NewCustomError(ErrTypeInvalidInput, "")
	ErrNotFound     = NewCustomError(ErrTypeNotFound, "")
	ErrConflict     = NewCustomError(ErrTypeConflict, "")
	ErrUnauthorized = NewCustomError(ErrTypeUnauthorized, "")
	ErrProvider     = NewCustomError(ErrTypeProvider, "")
	ErrUnsupported  = NewCustomError(ErrTypeUnsupported, "")
	ErrInternal     = NewCustomError(ErrTypeInternal, "")
)
