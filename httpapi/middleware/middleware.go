package middleware

import (
	"github.com/danielgtaylor/huma/v2"
)

type Middleware struct {
	api huma.API
}

func NewMiddleware(api huma.API) *Middleware {
	return &Middleware{api: api}
}
