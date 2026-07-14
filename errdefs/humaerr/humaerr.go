package humaerr

import (
	"github.com/danielgtaylor/huma/v2"
)

type ResponseError struct {
	Status  int      `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func (e *ResponseError) Error() string {
	return e.Message
}

func (e *ResponseError) GetStatus() int {
	return e.Status
}

var HumaErrorFunc = func(status int, message string, errs ...error) huma.StatusError {
	details := make([]string, len(errs))
	for i, err := range errs {
		details[i] = err.Error()
	}
	return &ResponseError{
		Status:  status,
		Message: message,
		Details: details,
	}
}
