package lyrics

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func isolate(t *testing.T) *Client {
	t.Helper()
	return New(Options{CacheDir: t.TempDir()})
}

func denyNetwork(t *testing.T, client *Client) {
	t.Helper()
	client.fetch = func(context.Context, Provider, Request) (*Lyrics, error) {
		t.Error("lyrics lookup reached the network")
		return nil, errNotFound
	}
}

func TestLookupCacheOnlyNeverFetches(t *testing.T) {
	client := isolate(t)
	denyNetwork(t, client)

	result, err := client.Lookup(context.Background(), Request{Artist: "Cynthia Harrell", Title: "Snake Eater", CacheOnly: true, Providers: []Provider{LRCLIB}})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if result.Found {
		t.Error("found = true, want false")
	}
}

func TestLookupPrefersSidecarOverNetwork(t *testing.T) {
	client := isolate(t)
	denyNetwork(t, client)

	dir := t.TempDir()
	track := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(filepath.Join(dir, "track.lrc"), []byte("[00:11.82]What a thrill"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := client.Lookup(context.Background(), Request{
		FileURL: "file://" + track,
	})
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if result.Source != "sidecar" || len(result.Synced) != 1 {
		t.Fatalf("got source %q with %d lines, want sidecar with 1", result.Source, len(result.Synced))
	}
}

func TestLookupCachesMissAndDoesNotRefetch(t *testing.T) {
	client := isolate(t)

	calls := 0
	client.fetch = func(context.Context, Provider, Request) (*Lyrics, error) {
		calls++
		return nil, errNotFound
	}

	req := Request{Artist: "Nobody", Title: "Unknown Song", Providers: []Provider{LRCLIB}}
	for range 2 {
		if _, err := client.Lookup(context.Background(), req); err != nil {
			t.Fatalf("Lookup() error = %v", err)
		}
	}
	if calls != 1 {
		t.Errorf("fetch called %d times, want 1", calls)
	}
}

func TestNegativeCacheExpires(t *testing.T) {
	client := isolate(t)

	key := cacheKey(Request{Artist: "Nobody", Title: "Unknown Song"})
	writeEntry(t, client, key, cacheEntry{Version: cacheVersion, Miss: true, Fetched: time.Now().Add(-negativeTTL - time.Hour).Unix()})

	if _, ok := client.loadCache(key); ok {
		t.Error("expired negative entry was served")
	}

	writeEntry(t, client, key, cacheEntry{Version: cacheVersion, Miss: true, Fetched: time.Now().Unix()})
	if _, ok := client.loadCache(key); !ok {
		t.Error("fresh negative entry was not served")
	}
}

func TestCacheKeyFoldsIdentity(t *testing.T) {
	base := cacheKey(Request{Artist: "Cynthia Harrell", Title: "Snake Eater", Album: "MGS3", Duration: 180 * time.Second})

	if got := cacheKey(Request{Artist: "  CYNTHIA   Harrell ", Title: "snake eater", Album: "mgs3", Duration: 180 * time.Second}); got != base {
		t.Error("key changed under case and whitespace folding")
	}
	if got := cacheKey(Request{Artist: "Cynthia Harrell", Title: "Way to Fall", Album: "MGS3", Duration: 180 * time.Second}); got == base {
		t.Error("different tracks share a key")
	}
}

func writeEntry(t *testing.T, client *Client, key string, entry cacheEntry) {
	t.Helper()
	path := client.cachePath(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLookupSeparatesDuration(t *testing.T) {
	client := isolate(t)
	calls := 0
	client.fetch = func(_ context.Context, _ Provider, req Request) (*Lyrics, error) {
		calls++
		if req.Duration == 180*time.Second {
			return nil, errNotFound
		}
		return &Lyrics{Plain: "long recording"}, nil
	}
	req := Request{Artist: "Artist", Title: "Track", Album: "Album", Duration: 180 * time.Second, Providers: []Provider{LRCLIB}}
	if _, err := client.Lookup(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	req.Duration = 240 * time.Second
	result, err := client.Lookup(context.Background(), req)
	if err != nil || !result.Found || calls != 2 {
		t.Fatalf("duration-specific lookup = %+v, %v, calls %d", result, err, calls)
	}
}

func TestLookupCoalescesBeforeRateLimit(t *testing.T) {
	client := isolate(t)
	client.rateBuckets[LRCLIB] = &rateBucket{tokens: 1, last: time.Now()}
	started := make(chan struct{})
	release := make(chan struct{})
	client.fetch = func(context.Context, Provider, Request) (*Lyrics, error) {
		close(started)
		<-release
		return &Lyrics{Plain: "line"}, nil
	}
	req := Request{Artist: "Artist", Title: "Track", Providers: []Provider{LRCLIB}}
	results := make(chan error, rateBurst+1)
	go func() {
		_, err := client.Lookup(context.Background(), req)
		results <- err
	}()
	<-started
	for range rateBurst {
		go func() {
			_, err := client.Lookup(context.Background(), req)
			results <- err
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(release)
	for range rateBurst + 1 {
		if err := <-results; err != nil {
			t.Fatalf("coalesced lookup failed: %v", err)
		}
	}
}

func TestLrclibPlainLyricsLineLimit(t *testing.T) {
	if _, err := decodeLrclib(lrclibResponse{PlainLyrics: strings.Repeat("line\n", MaxLines+1)}); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("plain response limit error = %v", err)
	}
}
