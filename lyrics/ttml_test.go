package lyrics

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTTMLVoicesAndBackingSurviveCache(t *testing.T) {
	client := isolate(t)
	document := `<tt xmlns:ttm="http://www.w3.org/ns/ttml#metadata"><head><metadata>
<ttm:agent xml:id="a" type="person"><ttm:name type="full">First singer</ttm:name></ttm:agent>
<ttm:agent xml:id="b" type="person"><ttm:name type="full">Second singer</ttm:name></ttm:agent>
</metadata></head><body><div>
<p begin="1" end="5" ttm:agent="a"><span begin="1" end="4">Lead</span><span ttm:role="x-bg"><span begin="2" end="5">Echo</span></span></p>
<p begin="3" end="6" ttm:agent="b"><span begin="3" end="6">Reply</span><span ttm:role="x-translation">Ignored translation</span></p>
</div></body></tt>`
	client.fetch = func(context.Context, Provider, Request) (*Lyrics, error) {
		return ParseTTML([]byte(document))
	}
	req := Request{Title: "Track", Artist: "Artist", Providers: []Provider{BetterLyrics}}
	want := []Line{
		{Start: 1, End: 4, Text: "Lead", Voice: "a", Group: 1, Words: []Word{{Start: 1, End: 4, Text: "Lead"}}},
		{Start: 2, End: 5, Text: "Echo", Voice: "a", Background: true, Group: 1, Words: []Word{{Start: 2, End: 5, Text: "Echo"}}},
		{Start: 3, End: 6, Text: "Reply", Voice: "b", Words: []Word{{Start: 3, End: 6, Text: "Reply"}}},
	}
	for _, cached := range []bool{false, true} {
		result, err := client.Lookup(context.Background(), req)
		if err != nil || result.Cached != cached || !reflect.DeepEqual(result.Synced, want) || result.Voices["a"].Name != "First singer" || result.Voices["b"].Type != "person" {
			t.Fatalf("cached=%v result=%+v err=%v", cached, result, err)
		}
	}
}

func TestTTMLWordTiming(t *testing.T) {
	result, err := ParseTTML([]byte(`<tt xmlns="http://www.w3.org/ns/ttml"><body><div begin="1.200"><p begin="1.200" end="3.0"><span end="1.600" begin="1.200">Hello</span> <span dur="1s" begin="1.750"><span>world</span></span><br/>again &amp; again</p></div></body></tt>`))
	want := []Word{{Start: 1.2, End: 1.6, Text: "Hello "}, {Start: 1.75, End: 2.75, Text: "world\nagain & again"}}
	if err != nil || len(result.Synced) != 1 || result.Synced[0].Start != 1.2 || result.Synced[0].Text != "Hello world\nagain & again" || !reflect.DeepEqual(result.Synced[0].Words, want) {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestTTMLLineAndPlainLyrics(t *testing.T) {
	for _, test := range []struct {
		name, body string
		time       Seconds
		synced     bool
	}{
		{"clock", `<p begin="01:02.500">Plain line</p>`, 62.5, true},
		{"milliseconds", `<p begin="1250ms">Plain line</p>`, 1.25, true},
		{"untimed", `<p>Plain line</p>`, 0, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseTTML([]byte("<tt><body><div>" + test.body + "</div></body></tt>"))
			if err != nil || result.Plain != "Plain line" || (len(result.Synced) > 0) != test.synced {
				t.Fatalf("result=%+v err=%v", result, err)
			}
			if test.synced && (result.Synced[0].Start != test.time || len(result.Synced[0].Words) != 0) {
				t.Fatalf("line=%+v", result.Synced[0])
			}
		})
	}
}

func TestTTMLIndentedWordsAndFragmentedText(t *testing.T) {
	result, err := ParseTTML([]byte("<tt><body><p>\n <span begin=\"6\">word</span>\n</p></body></tt>"))
	if err != nil || !reflect.DeepEqual(result.Synced[0], Line{Start: 6, Text: "word", Words: []Word{{Start: 6, Text: "word"}}}) {
		t.Fatalf("indented words=%+v err=%v", result, err)
	}
	result, err = ParseTTML([]byte("<tt><body><p begin=\"1\">" + strings.Repeat("a<!-- split -->", 20000) + "</p></body></tt>"))
	if err != nil || len(result.Plain) != 20000 {
		t.Fatalf("fragmented text length=%d err=%v", len(result.Plain), err)
	}
}

func TestTTMLRejectsInvalidAndExcessiveInput(t *testing.T) {
	for _, text := range []string{
		`<tt><body><p begin="NaN">bad</p></body></tt>`,
		`<tt><body><p begin="-1">bad</p></body></tt>`,
		`<tt><body><p begin="1:2:3:4">bad</p></body></tt>`,
		`<tt><body><p><span begin="1e308" dur="1e308">bad</span></p></body></tt>`,
		`<!DOCTYPE tt [<!ENTITY text "bad">]><tt><body><p>&text;</p></body></tt>`,
		`<tt><body><p>unclosed</body></tt>`,
		`<tt/><tt/>`,
		"<tt>" + strings.Repeat("<div>", 33) + strings.Repeat("</div>", 33) + "</tt>",
		"<tt><body>" + strings.Repeat("<p>line</p>", MaxLines+1) + "</body></tt>",
	} {
		if _, err := ParseTTML([]byte(text)); err == nil {
			t.Fatalf("invalid TTML accepted (%d bytes)", len(text))
		}
	}
}

func TestTTMLOversizedVoiceFallsBackToUnlabelledLyrics(t *testing.T) {
	id := strings.Repeat("a", 200000)
	result, err := ParseTTML([]byte(`<tt><body agent="` + id + `"><p begin="1" end="2">Text</p><p begin="2" end="3">More text</p></body></tt>`))
	if err != nil || len(result.Synced) != 2 || result.Synced[0].Voice != "" || result.Synced[1].Voice != "" {
		t.Fatalf("oversized voice result=%+v err=%v", result, err)
	}
}

func TestTTMLSidecar(t *testing.T) {
	client := isolate(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "track.ttml"), []byte(`<tt><body><p begin="2">From TTML</p></body></tt>`), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := client.Lookup(context.Background(), Request{FileURL: "file://" + filepath.Join(dir, "track.flac")})
	if err != nil || result.Source != Sidecar || len(result.Synced) != 1 || result.Synced[0].Text != "From TTML" {
		t.Fatalf("ttml sidecar: %+v, %v", result, err)
	}
}
