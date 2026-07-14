package httpapi

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type Option func(*huma.Config)

func WithDescription(d string) Option {
	return func(cfg *huma.Config) {
		cfg.Info.Description = d
	}
}

func WithURLEncodedForms() Option {
	return func(cfg *huma.Config) {
		cfg.Formats["application/x-www-form-urlencoded"] = urlEncodedFormat
		cfg.Formats["x-www-form-urlencoded"] = urlEncodedFormat
	}
}

func NewHumaConfig(title, version string, opts ...Option) huma.Config {
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)

	cfg := huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   title,
				Version: version,
			},
			Components: &huma.Components{
				Schemas: registry,
			},
		},
		OpenAPIPath:   "/openapi",
		DocsPath:      "/docs",
		SchemasPath:   "/schemas",
		Formats:       huma.DefaultFormats,
		DefaultFormat: "application/json",
	}

	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func DocsHandler(title string) http.HandlerFunc {
	page := fmt.Sprintf(`<!doctype html>
<html>
	<head>
		<title>%s</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
	</head>
	<body>
		<script id="api-reference" data-url="/openapi.json"></script>
		<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
</html>`, title)

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

type EmptyInput struct{}

type DeletedResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}
