package lyrics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const wordLyricsfile = `version: "1.0"
metadata:
  title: Fixture
  artist: Fixture
lines:
  - text: Hello world
    start_ms: 1200
    end_ms: 2800
    words:
      - text: "Hello "
        start_ms: 1200
        end_ms: 1900
      - text: world
        start_ms: 1900
        end_ms: 2800
plain: |
  Hello world

  Another verse
`

func TestLyricsfileLookupPreservesWordsThroughCache(t *testing.T) {
	client := isolate(t)
	payload := lrclibResponse{Lyricsfile: wordLyricsfile, SyncedLyrics: "[00:01.20]Legacy line"}
	client.fetch = func(context.Context, Provider, Request) (*Lyrics, error) {
		return decodeLrclib(payload)
	}
	req := Request{Artist: "Fixture", Title: "Fixture", Providers: []Provider{LRCLIB}}
	writeEntry(t, client, cacheKey(req), cacheEntry{Version: cacheVersion - 1, Fetched: time.Now().Unix(), Lyrics: Lyrics{Plain: "old line-only result"}})
	for i := 0; i < 2; i++ {
		result, err := client.Lookup(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Found || result.Cached != (i == 1) || result.Source != "lrclib" || result.Plain != "Hello world\n\nAnother verse\n" {
			t.Fatalf("lookup %d: %+v", i, result)
		}
		assertLines(t, result.Synced, []Line{{Start: 1.2, Text: "Hello world", Words: []Word{{Start: 1.2, End: 1.9, Text: "Hello "}, {Start: 1.9, End: 2.8, Text: "world"}}}})
		denyNetwork(t, client)
		req.CacheOnly = true
	}
}

func TestLyricsfileFallbackToLRC(t *testing.T) {
	for _, document := range []string{
		"[broken yaml",
		strings.Replace(wordLyricsfile, `version: "1.0"`, `version: "2.0"`, 1),
		wordLyricsfile + "\n---\nversion: '1.0'\n",
		wordLyricsfile + "\nplain: duplicate\n",
		wordLyricsfile + "\nextra: &alias [*alias]\n",
		wordLyricsfile + "\nextra: !custom value\n",
		strings.Replace(wordLyricsfile, "start_ms: 1200", "start_ms: -1", 1),
		strings.Replace(wordLyricsfile, "start_ms: 1200", "start_ms: 1.5", 1),
		strings.Replace(wordLyricsfile, "start_ms: 1200", "missing_start: 1200", 1),
		strings.Replace(wordLyricsfile, "end_ms: 2800", "end_ms: 100", 1),
	} {
		payload := lrclibResponse{Lyricsfile: document, SyncedLyrics: "[00:02]Still readable"}
		result, err := decodeLrclib(payload)
		if err != nil {
			t.Fatal(err)
		}
		assertLines(t, result.Synced, []Line{{Start: 2, Text: "Still readable"}})
	}
}

func TestLyricsfilePlainAndInstrumental(t *testing.T) {
	for _, document := range []string{
		"version: '1.0'\nmetadata: {}\nplain: |\n  First\n\n  Last\n",
		"version: '1.0'\nmetadata: {instrumental: true}\n",
	} {
		payload := lrclibResponse{Lyricsfile: document}
		result, err := decodeLrclib(payload)
		if err != nil || result.Empty() || len(result.Synced) != 0 {
			t.Fatalf("Lyricsfile-only response: %+v, %v", result, err)
		}
		if !result.Instrumental && result.Plain != "First\n\nLast\n" {
			t.Fatalf("plain lyrics changed: %q", result.Plain)
		}
	}
}

func TestLyricsfileSidecarAndFallback(t *testing.T) {
	client := isolate(t)
	denyNetwork(t, client)
	dir := t.TempDir()
	path := filepath.Join(dir, "track.lyricsfile.yaml")
	if err := os.WriteFile(path, []byte(wordLyricsfile), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "track.lrc"), []byte("[00:02]LRC fallback"), 0o600); err != nil {
		t.Fatal(err)
	}
	req := Request{FileURL: "file://" + filepath.Join(dir, "track.flac")}
	result, err := client.Lookup(context.Background(), req)
	if err != nil || result.Source != "sidecar" || len(result.Synced) != 1 || len(result.Synced[0].Words) != 2 {
		t.Fatalf("word sidecar: %+v, %v", result, err)
	}
	if err := os.WriteFile(path, []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err = client.Lookup(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, result.Synced, []Line{{Start: 2, Text: "LRC fallback"}})
}

func TestLyricsfileIncompleteWordsPreserveLine(t *testing.T) {
	for _, document := range []string{
		strings.Replace(wordLyricsfile, "text: world", "missing_text: world", 1),
		strings.Replace(wordLyricsfile, "text: world", "text: there", 1),
		strings.Replace(wordLyricsfile, "start_ms: 1900", "start_ms: -1", 1),
		strings.Replace(wordLyricsfile, "start_ms: 1900", "missing_start: 1900", 1),
	} {
		result, err := ParseLyricsfile([]byte(document))
		if err != nil {
			t.Fatal(err)
		}
		assertLines(t, result.Synced, []Line{{Start: 1.2, Text: "Hello world"}})
	}
}

func TestLyricsfileLimits(t *testing.T) {
	var fields strings.Builder
	for i := 0; i < 65; i++ {
		fmt.Fprintf(&fields, "field%d: value\n", i)
	}
	for _, document := range []string{
		fields.String(),
		"version: '1.0'\nlines:\n" + strings.Repeat("- {text: x, start_ms: 0}\n", MaxLines+1),
		"version: '1.0'\nlines:\n- text: x\n  start_ms: 0\n  words:\n" + strings.Repeat("  - {text: x, start_ms: 0}\n", MaxWords+1),
		"version: '1.0'\nplain: |\n" + strings.Repeat("  line\n", MaxLines+1),
		"version: '1.0'\nextra: " + strings.Repeat("[", 20) + strings.Repeat("]", 20),
	} {
		if _, err := ParseLyricsfile([]byte(document)); !errors.Is(err, ErrTooLarge) {
			t.Fatalf("Lyricsfile limit error = %v", err)
		}
	}
}
