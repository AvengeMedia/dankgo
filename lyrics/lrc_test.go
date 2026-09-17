package lyrics

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseTimestamps(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Line
	}{
		{
			name:  "minutes and seconds",
			input: "[00:11]What a thrill",
			want:  []Line{{Start: 11, Text: "What a thrill"}},
		},
		{
			name:  "centiseconds",
			input: "[01:05.30]Crime, it's the way",
			want:  []Line{{Start: 65.30, Text: "Crime, it's the way"}},
		},
		{
			name:  "milliseconds",
			input: "[00:11.820]What a thrill",
			want:  []Line{{Start: 11.82, Text: "What a thrill"}},
		},
		{
			name:  "comma fraction",
			input: "[00:11,82]What a thrill",
			want:  []Line{{Start: 11.82, Text: "What a thrill"}},
		},
		{
			name:  "hours",
			input: "[01:02:03.50]Late",
			want:  []Line{{Start: 3723.50, Text: "Late"}},
		},
		{
			name:  "one line repeated at several times",
			input: "[00:11.82][01:05.30][02:00.00](Snake Eater)",
			want: []Line{
				{Start: 11.82, Text: "(Snake Eater)"},
				{Start: 65.30, Text: "(Snake Eater)"},
				{Start: 120, Text: "(Snake Eater)"},
			},
		},
		{
			name:  "empty text is an intentional blank",
			input: "[00:05.00]\n[00:07.00]After",
			want: []Line{
				{Start: 5, Text: ""},
				{Start: 7, Text: "After"},
			},
		},
		{
			name:  "unsorted input is sorted",
			input: "[00:20.00]Second\n[00:10.00]First",
			want: []Line{
				{Start: 10, Text: "First"},
				{Start: 20, Text: "Second"},
			},
		},
		{
			name:  "duplicate timestamps keep source order",
			input: "[00:10.00]Lead\n[00:10.00]Backing",
			want: []Line{
				{Start: 10, Text: "Lead"},
				{Start: 10, Text: "Backing"},
			},
		},
		{
			name:  "enhanced word timings are preserved",
			input: "[00:11.82]<00:11.82>What <00:12.10>a <00:12.50>thrill",
			want:  []Line{{Start: 11.82, Text: "What a thrill", Words: []Word{{Start: 11.82, End: 12.10, Text: "What "}, {Start: 12.10, End: 12.50, Text: "a "}, {Start: 12.50, Text: "thrill"}}}},
		},
		{
			name:  "word offsets and repetitions retain spacing and unicode",
			input: "[offset:500]\n[00:10][00:20] <00:10>Hello <00:10.250>世<00:10.750>界<00:11> ",
			want: []Line{
				{Start: 9.5, Text: "Hello 世界", Words: []Word{{Start: 9.5, End: 9.75, Text: "Hello "}, {Start: 9.75, End: 10.25, Text: "世"}, {Start: 10.25, End: 10.5, Text: "界"}}},
				{Start: 19.5, Text: "Hello 世界", Words: []Word{{Start: 19.5, End: 19.75, Text: "Hello "}, {Start: 19.75, End: 20.25, Text: "世"}, {Start: 20.25, End: 20.5, Text: "界"}}},
			},
		},
		{
			name:  "invalid word order falls back to line timing",
			input: "[00:10]<00:12>Hello <00:11>world",
			want:  []Line{{Start: 10, Text: "Hello world"}},
		},
		{
			name:  "metadata tags are skipped",
			input: "[ar:Cynthia Harrell]\n[ti:Snake Eater]\n[length:03:24]\n[00:11.82]What a thrill",
			want:  []Line{{Start: 11.82, Text: "What a thrill"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseLRC([]byte(test.input))
			if err != nil {
				t.Fatalf("ParseLRC() error = %v", err)
			}
			assertLines(t, parsed.Synced, test.want)
		})
	}
}

func TestParseOffsetSign(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Seconds
	}{
		{name: "positive shifts earlier", input: "[offset:+500]\n[00:10.00]Line", want: 9.5},
		{name: "unsigned is positive", input: "[offset:500]\n[00:10.00]Line", want: 9.5},
		{name: "negative shifts later", input: "[offset:-500]\n[00:10.00]Line", want: 10.5},
		{name: "clamped at zero", input: "[offset:+5000]\n[00:01.00]Line", want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := ParseLRC([]byte(test.input))
			if err != nil {
				t.Fatalf("ParseLRC() error = %v", err)
			}
			got := parsed.Synced
			if len(got) != 1 {
				t.Fatalf("got %d lines, want 1", len(got))
			}
			if !nearly(got[0].Start, test.want) {
				t.Errorf("time = %v, want %v", got[0].Start, test.want)
			}
		})
	}
}

func TestParsePlainFallback(t *testing.T) {
	parsed, err := ParseLRC([]byte("What a thrill\nWith darkness and silence"))
	if err != nil {
		t.Fatalf("ParseLRC() error = %v", err)
	}
	if len(parsed.Synced) != 0 {
		t.Errorf("got %d synced lines, want 0", len(parsed.Synced))
	}
	if parsed.Plain != "What a thrill\nWith darkness and silence" {
		t.Errorf("plain = %q", parsed.Plain)
	}
}

func TestParseMetadataOnlyYieldsNothing(t *testing.T) {
	parsed, err := ParseLRC([]byte("[ar:Artist]\n[ti:Title]\n[by:Someone]"))
	if err != nil {
		t.Fatalf("ParseLRC() error = %v", err)
	}
	if !parsed.Empty() {
		t.Errorf("got %+v, want nothing", parsed)
	}
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  error
	}{
		{name: "invalid utf-8", input: []byte{0x5b, 0x30, 0x30, 0xff, 0xfe, 0x41}, want: ErrEncoding},
		{name: "oversize source", input: []byte(strings.Repeat("a", MaxSourceBytes+1)), want: ErrTooLarge},
		{name: "too many lines", input: []byte(strings.Repeat("[00:01.00]x\n", MaxLines+1)), want: ErrTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseLRC(test.input); !errors.Is(err, test.want) {
				t.Errorf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestParseStripsBOM(t *testing.T) {
	parsed, err := ParseLRC(append(utf8BOM, []byte("[00:11.82]What a thrill")...))
	if err != nil {
		t.Fatalf("ParseLRC() error = %v", err)
	}
	assertLines(t, parsed.Synced, []Line{{Start: 11.82, Text: "What a thrill"}})
}

func assertLines(t *testing.T, got, want []Line) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !nearly(got[i].Start, want[i].Start) || got[i].Text != want[i].Text {
			t.Errorf("line %d = {%v %q}, want {%v %q}", i, got[i].Start, got[i].Text, want[i].Start, want[i].Text)
		}
		if len(got[i].Words) != len(want[i].Words) {
			t.Errorf("line %d words = %+v, want %+v", i, got[i].Words, want[i].Words)
			continue
		}
		for j, word := range want[i].Words {
			if !nearly(got[i].Words[j].Start, word.Start) || !nearly(got[i].Words[j].End, word.End) || got[i].Words[j].Text != word.Text {
				t.Errorf("line %d word %d = %+v, want %+v", i, j, got[i].Words[j], word)
			}
		}
	}
}

func nearly(a, b Seconds) bool {
	diff := a - b
	return diff < 0.0005 && diff > -0.0005
}

func TestParseRejectsRepeatedWordOverflow(t *testing.T) {
	text := "[00:00][00:1e308]<00:00>x<00:1" + strings.Repeat("0", 308) + ">"
	if _, err := ParseLRC([]byte(text)); err == nil {
		t.Fatal("repeated word timestamp overflow was accepted")
	}
}

func TestParseNonfiniteTimesRemainSerializable(t *testing.T) {
	for _, input := range []string{"[NaN:01]Line", "[00:+Inf]Line", "[1e308:01]Line", "[offset:NaN]\n[00:01]Line", "[offset:-Inf]\n[00:01]Line", "[offset:-1.7976931348623157e308]\n[0:1.7976931348623157e308]Line", "[offset:-1.7976931348623157e308]\n[0:0]<0:1.7976931348623157e308>Line"} {
		parsed, err := ParseLRC([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := json.Marshal(parsed.Synced); err != nil {
			t.Fatalf("%q produced unserializable times: %v", input, err)
		}
	}
}

func TestParseBoundsExpandedLyrics(t *testing.T) {
	for _, input := range []string{
		strings.Repeat("[00:01]", MaxLines) + strings.Repeat("x", 1024),
		strings.Repeat("plain line\n", MaxLines+1),
		"[00:00]" + strings.Repeat("<00:00>x", MaxWords+1),
		strings.Repeat("[00:00]", 200) + strings.Repeat("<00:00>x", 100),
	} {
		if _, err := ParseLRC([]byte(input)); !errors.Is(err, ErrTooLarge) {
			t.Fatalf("expanded lyrics error = %v, want ErrTooLarge", err)
		}
	}
}
