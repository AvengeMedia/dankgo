package lyrics

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	stub(client, func(context.Context, Provider, Request) (*Lyrics, error) {
		return decodeLrclib(payload)
	})
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

func TestAsLyricsfileRoundTrip(t *testing.T) {
	source := &Lyrics{
		Title:      "Fixture",
		Artist:     "Fixture",
		Album:      "Fixture",
		DurationMS: 30000,
		Synced: []Line{{
			Start: 1.2, End: 2.8, Text: "Hello world",
			Words: []Word{{Start: 1.2, End: 1.9, Text: "Hello "}, {Start: 1.9, End: 2.8, Text: "world"}},
		}},
		Plain: "Hello world",
	}
	out, err := source.AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"title: Fixture", "artist: Fixture", "album: Fixture", "duration_ms: 30000", "start_ms: 1200", "end_ms: 2800", "words:"} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	again, err := ParseLyricsfile(out)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, again.Synced, source.Synced)
	if len(again.Synced) != 1 || !nearly(again.Synced[0].End, source.Synced[0].End) {
		t.Fatalf("line end = %+v, want %+v", again.Synced, source.Synced)
	}
	if again.Title != source.Title || again.Artist != source.Artist || again.Album != source.Album ||
		again.DurationMS != source.DurationMS || again.Plain != source.Plain {
		t.Fatalf("document changed: %+v != %+v", again, source)
	}
	// Clear the raw passthrough so the second pass re-encodes instead of
	// echoing the input bytes back.
	again.Source, again.raw = SourceUnknown, nil
	second, err := again.AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != string(out) {
		t.Fatalf("second marshal differs:\nfirst:\n%s\nsecond:\n%s", out, second)
	}
}

func TestAsLyricsfileSortsOutput(t *testing.T) {
	// ParseLyricsfile sorts on read, so a round trip would hide a missing
	// writer-side sort: assert on the marshalled bytes themselves.
	source := &Lyrics{
		Title:  "T",
		Artist: "A",
		Synced: []Line{{Start: 20, Text: "Second"}, {Start: 10, Text: "First"}},
		Plain:  "First\nSecond",
	}
	out, err := source.AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(out), "Second") < strings.Index(string(out), "First") {
		t.Fatalf("lines are not sorted by start:\n%s", out)
	}
	back, err := ParseLyricsfile(out)
	if err != nil {
		t.Fatal(err)
	}
	assertLines(t, back.Synced, []Line{{Start: 10, Text: "First"}, {Start: 20, Text: "Second"}})
}

func TestAsLyricsfileReturnsSourceVerbatim(t *testing.T) {
	parsed, err := ParseLyricsfile([]byte(wordLyricsfile))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Source != SourceLyricsfile || parsed.Title != "Fixture" || parsed.Artist != "Fixture" {
		t.Fatalf("identity dropped: %+v", parsed)
	}
	out, err := parsed.AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != wordLyricsfile {
		t.Fatalf("source was rewritten:\n%s", out)
	}
}

func TestAsLyricsfileFromAnySource(t *testing.T) {
	for _, test := range []struct {
		name   string
		parse  Parser
		source string
	}{
		{"line-synced lrc", ParseLRC, "[00:01.20]Hello world\n[00:03.00]Second line"},
		{"word-synced lrc", ParseLRC, "[00:01.20]<00:01.20>Hel<00:01.50>lo world"},
		{"plain-only lrc", ParseLRC, "no timestamps here\njust text"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source, err := test.parse([]byte(test.source))
			if err != nil {
				t.Fatal(err)
			}
			source.Title, source.Artist = "T", "A"
			marshaled, err := source.AsLyricsfile()
			if err != nil {
				t.Fatal(err)
			}
			back, err := ParseLyricsfile(marshaled)
			if err != nil {
				t.Fatal(err)
			}
			assertLines(t, back.Synced, source.Synced)
			if back.Plain != source.Plain {
				t.Fatalf("plain = %q, want %q", back.Plain, source.Plain)
			}
		})
	}
}

func TestAsLyricsfileInstrumentalAndEmpty(t *testing.T) {
	instrumental, err := (&Lyrics{Instrumental: true, Title: "T", Artist: "A"}).AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"lines", "plain"} {
		if strings.Contains(string(instrumental), unwanted) {
			t.Fatalf("instrumental document contains %q:\n%s", unwanted, instrumental)
		}
	}
	back, err := ParseLyricsfile(instrumental)
	if err != nil || !back.Instrumental {
		t.Fatalf("instrumental round trip: %+v, %v", back, err)
	}
	empty, err := (&Lyrics{Title: "T", Artist: "A"}).AsLyricsfile()
	if err != nil {
		t.Fatal(err)
	}
	if back, err = ParseLyricsfile(empty); err != nil || !back.Empty() {
		t.Fatalf("empty round trip: %+v, %v", back, err)
	}
}

func TestAsLyricsfileDropsBrokenWords(t *testing.T) {
	for _, line := range []Line{
		{Start: 1, End: 2, Text: "Hello world", Words: []Word{{Start: 1, Text: "Hello "}, {Start: 1.5, Text: "there"}}},
		{Start: 1, End: 2, Text: "Hello world", Words: []Word{{Start: 1, Text: "Hello "}, {Start: -1, Text: "world"}}},
	} {
		out, err := (&Lyrics{Title: "T", Artist: "A", Synced: []Line{line}, Plain: "Hello world"}).AsLyricsfile()
		if err != nil {
			t.Fatal(err)
		}
		back, err := ParseLyricsfile(out)
		if err != nil {
			t.Fatal(err)
		}
		assertLines(t, back.Synced, []Line{{Start: 1, End: 2, Text: "Hello world"}})
	}
}

func TestAsLyricsfileInvalid(t *testing.T) {
	var missing *Lyrics
	if _, err := missing.AsLyricsfile(); !errors.Is(err, errLyricsfile) {
		t.Fatalf("nil receiver error = %v", err)
	}
	if _, err := (&Lyrics{}).AsLyricsfile(); !errors.Is(err, errLyricsfile) {
		t.Fatalf("missing identity error = %v", err)
	}
	if _, err := (&Lyrics{Title: "T", Artist: "A", DurationMS: -1}).AsLyricsfile(); !errors.Is(err, errLyricsfile) {
		t.Fatalf("negative duration error = %v", err)
	}
	for _, line := range []Line{
		{Start: -1, Text: "x"},
		{Start: 2, End: 1, Text: "x"},
		{Start: Seconds(math.Inf(1)), Text: "x"},
	} {
		if _, err := (&Lyrics{Title: "T", Artist: "A", Synced: []Line{line}}).AsLyricsfile(); !errors.Is(err, errLyricsfile) {
			t.Fatalf("line %+v error = %v", line, err)
		}
	}
	tooManyLines := make([]Line, MaxLines+1)
	if _, err := (&Lyrics{Title: "T", Artist: "A", Synced: tooManyLines}).AsLyricsfile(); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("line limit error = %v", err)
	}
	// Plain short enough that the plain-count guard cannot mask the
	// Synced-count guard: only len(Synced) > MaxLines can reject this.
	if _, err := (&Lyrics{Title: "T", Artist: "A", Synced: tooManyLines, Plain: "short"}).AsLyricsfile(); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("line limit error with plain set = %v", err)
	}
	oneLine := Line{Start: 1, Text: strings.Repeat("x", MaxWords+1)}
	oneLine.Words = make([]Word, MaxWords+1)
	if _, err := (&Lyrics{Title: "T", Artist: "A", Synced: []Line{oneLine}}).AsLyricsfile(); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("word limit error = %v", err)
	}
	if _, err := (&Lyrics{Title: "T", Artist: "A", Plain: strings.Repeat("line\n", MaxLines+1)}).AsLyricsfile(); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("plain limit error = %v", err)
	}
}
