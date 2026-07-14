package ipc

import (
	"context"
	"strings"
)

type prefixRoute struct {
	prefix  string
	handler Handler
}

type Mux struct {
	exact    map[string]Handler
	prefixes []prefixRoute
}

func NewMux() *Mux {
	return &Mux{exact: make(map[string]Handler)}
}

func (m *Mux) Handle(method string, h Handler) {
	m.exact[method] = h
}

func (m *Mux) HandlePrefix(prefix string, h Handler) {
	m.prefixes = append(m.prefixes, prefixRoute{prefix: prefix, handler: h})
}

func (m *Mux) ServeIPC(ctx context.Context, w *ConnWriter, req Request, sub *Subscriber) {
	if h, ok := m.exact[req.Method]; ok {
		h(ctx, w, req, sub)
		return
	}
	for _, r := range m.prefixes {
		if strings.HasPrefix(req.Method, r.prefix) {
			r.handler(ctx, w, req, sub)
			return
		}
	}
	RespondError(w, req.ID, "unknown method: "+req.Method)
}
