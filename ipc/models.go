package ipc

import (
	"github.com/AvengeMedia/dankgo/log"
)

type Capabilities struct {
	APIVersion   int      `json:"apiVersion"`
	Capabilities []string `json:"capabilities"`
}

type Request struct {
	ID     int            `json:"id,omitempty"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type Response[T any] struct {
	ID     int    `json:"id,omitempty"`
	Result *T     `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func RespondError(w *ConnWriter, id int, msg string) {
	log.Errorf("ipc error: id=%d method-error=%s", id, msg)
	_ = w.WriteResponse(Response[any]{ID: id, Error: msg})
}

func Respond[T any](w *ConnWriter, id int, result T) {
	_ = w.WriteResponse(Response[T]{ID: id, Result: &result})
}
