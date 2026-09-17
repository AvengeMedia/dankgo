package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const defaultUserAgent = "dankgo-lyrics (+https://github.com/AvengeMedia/dankgo)"

var (
	errNotFound = errors.New("lyrics not found")
	errUncached = fmt.Errorf("%w: not cached by provider", errNotFound)
)

// StatusError is a non-2xx provider response that is not a plain "no such track".
type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("HTTP %d", e.Code)
}

// A new provider is a fetch method plus a row here.
var fetchers = map[Provider]func(*Client, context.Context, Request) (*Lyrics, error){
	LRCLIB:       (*Client).fetchLrclib,
	BetterLyrics: (*Client).fetchBetterLyrics,
	LyricsPlus:   (*Client).fetchLyricsPlus,
	Unison:       (*Client).fetchUnison,
}

// wordProviders are worth holding a line-synced winner for. lrclib has word sync too rarely.
var wordProviders = map[Provider]bool{
	BetterLyrics: true,
	Unison:       true,
	LyricsPlus:   true,
}

func (c *Client) fetchProvider(ctx context.Context, provider Provider, req Request) (*Lyrics, error) {
	return fetchers[provider](c, ctx, req)
}

func newHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	return &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			DialContext:         dialer.DialContext,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		CheckRedirect: sameOrigin,
	}
}

func sameOrigin(req *http.Request, via []*http.Request) error {
	origin := via[0].URL
	if len(via) >= 5 || req.URL.Scheme != origin.Scheme || req.URL.Host != origin.Host {
		return errors.New("refusing lyrics redirect")
	}
	return nil
}

func (c *Client) get(ctx context.Context, endpoint string, query url.Values, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, classifyStatus(resp.StatusCode)
	}
	return readBounded(resp.Body)
}

// betterlyrics answers 401 for a track its server has not cached yet, which another
// client can change at any time. Only a transient status is worth reporting as an error.
func classifyStatus(code int) error {
	switch code {
	case http.StatusUnauthorized:
		return errUncached
	case http.StatusNotFound, http.StatusForbidden, http.StatusGone:
		return errNotFound
	}
	return &StatusError{Code: code}
}

func readBounded(stream io.Reader) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(stream, MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > MaxSourceBytes {
		return nil, ErrTooLarge
	}
	return body, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, query url.Values, timeout time.Duration, payload any) error {
	body, err := c.get(ctx, endpoint, query, timeout)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, payload); err != nil {
		return fmt.Errorf("invalid lyrics response: %w", err)
	}
	return nil
}
