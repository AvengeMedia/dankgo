package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AvengeMedia/dankgo/httpapi"
	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type greetOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func newTestRouter(t *testing.T) chi.Router {
	t.Helper()
	return httpapi.NewRouter(httpapi.RouterConfig{Title: "Test API", Version: "1.0.0"}, func(api huma.API, r chi.Router) {
		huma.Get(api, "/greet", func(ctx context.Context, input *httpapi.EmptyInput) (*greetOutput, error) {
			out := &greetOutput{}
			out.Body.Message = "hello"
			return out, nil
		})
		huma.Get(api, "/boom", func(ctx context.Context, input *httpapi.EmptyInput) (*greetOutput, error) {
			panic("kaboom")
		})
	})
}

func TestRouterHealthAndHandler(t *testing.T) {
	router := newTestRouter(t)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/greet", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"message":"hello"}`, rec.Body.String())

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "api-reference")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecovererReturnsErrorEnvelope(t *testing.T) {
	router := newTestRouter(t)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "Internal server error")
}

func TestServerServeAndShutdown(t *testing.T) {
	srv := httpapi.NewServer("127.0.0.1:0", newTestRouter(t))
	require.NoError(t, srv.Listen())

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx) }()

	resp, err := http.Get("http://" + srv.Addr() + "/health")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cancel()
	require.NoError(t, <-done)
}
