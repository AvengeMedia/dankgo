package errdefs

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinStatusAndString(t *testing.T) {
	tests := []struct {
		err    *CustomError
		name   string
		status int
	}{
		{ErrInvalidInput, "ErrInvalidInput", http.StatusBadRequest},
		{ErrNotFound, "ErrNotFound", http.StatusNotFound},
		{ErrConflict, "ErrConflict", http.StatusConflict},
		{ErrUnauthorized, "ErrUnauthorized", http.StatusUnauthorized},
		{ErrProvider, "ErrProvider", http.StatusInternalServerError},
		{ErrUnsupported, "ErrUnsupported", http.StatusNotImplemented},
		{ErrInternal, "ErrInternal", http.StatusInternalServerError},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.name, tc.err.Type.String())
		assert.Equal(t, tc.status, tc.err.GetStatus())
	}
}

func TestRegisterAppError(t *testing.T) {
	const ErrTypeTeapot = AppErrorBase + 1
	Register(ErrTypeTeapot, "ErrTeapot", http.StatusTeapot)

	err := NewCustomError(ErrTypeTeapot, "short and stout")
	assert.Equal(t, "ErrTeapot: short and stout", err.Error())
	assert.Equal(t, http.StatusTeapot, err.GetStatus())
}

func TestUnknownType(t *testing.T) {
	err := NewCustomError(ErrorType(999), "mystery")
	assert.Equal(t, "ErrUnknown: mystery", err.Error())
	assert.Equal(t, http.StatusInternalServerError, err.GetStatus())
}

func TestIsMatchesByType(t *testing.T) {
	wrapped := fmt.Errorf("outer: %w", NewCustomError(ErrTypeNotFound, "thing missing"))
	assert.True(t, errors.Is(wrapped, ErrNotFound))
	assert.False(t, errors.Is(wrapped, ErrConflict))
	assert.False(t, errors.Is(errors.New("plain"), ErrNotFound))
}
