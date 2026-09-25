package lyrics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("{}"))
	}))
	defer elsewhere.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("case") {
		case "agent":
			w.Write([]byte(r.UserAgent()))
		case "query":
			w.Write([]byte(r.URL.RawQuery))
		case "large":
			w.Write([]byte(strings.Repeat("x", MaxSourceBytes+1)))
		case "redirect":
			http.Redirect(w, r, elsewhere.URL, http.StatusFound)
		case "unavailable":
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	client := New(Options{CacheDir: t.TempDir(), UserAgent: "test-agent"})
	get := func(name string) ([]byte, error) {
		return client.get(context.Background(), server.URL, url.Values{"case": {name}}, time.Second)
	}

	if body, err := get("agent"); err != nil || string(body) != "test-agent" {
		t.Errorf("user agent = %q, %v", body, err)
	}
	if body, err := client.get(context.Background(), server.URL, url.Values{"case": {"query"}, "q": {"a - b"}}, time.Second); err != nil || !strings.Contains(string(body), "q=a%20-%20b") {
		t.Errorf("query = %q, %v", body, err)
	}
	if _, err := get("large"); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversized body error = %v", err)
	}
	if _, err := get("redirect"); err == nil {
		t.Error("followed a redirect to another host")
	}
	var status *StatusError
	if _, err := get("unavailable"); !errors.As(err, &status) || status.Code != http.StatusServiceUnavailable {
		t.Errorf("503 error = %v", err)
	}
}
