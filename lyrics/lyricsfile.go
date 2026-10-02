package lyrics

import (
	"errors"
	"io"
	"math"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

var errLyricsfile = errors.New("invalid Lyricsfile document")

type lyricsfileDocument struct {
	Version  string `yaml:"version"`
	Metadata struct {
		Title        string `yaml:"title"`
		Artist       string `yaml:"artist"`
		Album        string `yaml:"album,omitempty"`
		DurationMS   int64  `yaml:"duration_ms,omitempty"`
		Instrumental bool   `yaml:"instrumental,omitempty"`
	} `yaml:"metadata"`
	Lines []lyricsfileLine `yaml:"lines,omitempty"`
	Plain string           `yaml:"plain,omitempty"`
}

type lyricsfileLine struct {
	lyricsfileSpan `yaml:",inline"`
	Words          []lyricsfileSpan `yaml:"words,omitempty"`
}

type lyricsfileSpan struct {
	Text    string `yaml:"text"`
	StartMS *int64 `yaml:"start_ms,omitempty"`
	EndMS   *int64 `yaml:"end_ms,omitempty"`
}

func (s lyricsfileSpan) timing() (start, end Seconds, ok bool) {
	if s.StartMS == nil || *s.StartMS < 0 {
		return 0, 0, false
	}
	if s.EndMS == nil {
		return milliseconds(*s.StartMS), 0, true
	}
	return milliseconds(*s.StartMS), milliseconds(*s.EndMS), *s.EndMS >= *s.StartMS
}

// ParseLyricsfile accepts Lyricsfile version 1.0 only.
func ParseLyricsfile(data []byte) (*Lyrics, error) {
	text, err := decode(data)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(text))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return nil, err
	}
	if len(node.Content) != 1 || node.Content[0].Kind != yaml.MappingNode {
		return nil, errLyricsfile
	}
	if err := validateLyricsfileNode(&node, 0); err != nil {
		return nil, err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, errLyricsfile
	}
	var document lyricsfileDocument
	if err := node.Decode(&document); err != nil {
		return nil, err
	}
	if document.Version != "1.0" {
		return nil, errLyricsfile
	}
	if document.Metadata.Instrumental {
		return &Lyrics{
			Instrumental: true,
			Title:        document.Metadata.Title,
			Artist:       document.Metadata.Artist,
			Album:        document.Metadata.Album,
			DurationMS:   document.Metadata.DurationMS,
			Source:       SourceLyricsfile,
			raw:          data,
		}, nil
	}
	if len(document.Lines) > MaxLines || strings.Count(document.Plain, "\n")+1 > MaxLines {
		return nil, ErrTooLarge
	}
	result := &Lyrics{
		Plain:      document.Plain,
		Title:      document.Metadata.Title,
		Artist:     document.Metadata.Artist,
		Album:      document.Metadata.Album,
		DurationMS: document.Metadata.DurationMS,
		Source:     SourceLyricsfile,
		raw:        data,
	}
	wordCount := 0
	for _, source := range document.Lines {
		start, end, ok := source.timing()
		if !ok {
			return nil, errLyricsfile
		}
		wordCount += len(source.Words)
		if wordCount > MaxWords {
			return nil, ErrTooLarge
		}
		line := Line{Start: start, End: end, Text: source.Text}
		words := lyricsfileWords(source.Words)
		wordsText := joinWords(words)
		if line.Text == "" {
			line.Text = wordsText
		}
		if wordsText == line.Text && strings.TrimSpace(wordsText) != "" {
			line.Words = words
			for _, word := range words {
				line.Start = min(line.Start, word.Start)
			}
		}
		result.Synced = append(result.Synced, line)
	}
	sortLines(result.Synced)
	if result.Plain == "" {
		plain := make([]string, len(result.Synced))
		for i, line := range result.Synced {
			plain[i] = line.Text
		}
		result.Plain = strings.Join(plain, "\n")
	}
	return result, nil
}

func lyricsfileWords(source []lyricsfileSpan) []Word {
	words := make([]Word, 0, len(source))
	for _, span := range source {
		start, end, ok := span.timing()
		if !ok {
			return nil
		}
		words = append(words, Word{Start: start, End: end, Text: span.Text})
	}
	return words
}

// AsLyricsfile encodes l as Lyricsfile version 1.0. A Lyricsfile source is
// returned verbatim; every other source needs Title and Artist, which the
// format requires. Voices, background and group have no Lyricsfile
// representation and are dropped.
func (l *Lyrics) AsLyricsfile() ([]byte, error) {
	if l == nil {
		return nil, errLyricsfile
	}
	if l.Source == SourceLyricsfile && l.raw != nil {
		return l.raw, nil
	}
	if l.Title == "" || l.Artist == "" || l.DurationMS < 0 {
		return nil, errLyricsfile
	}
	document := lyricsfileDocument{Version: "1.0"}
	document.Metadata.Title = l.Title
	document.Metadata.Artist = l.Artist
	document.Metadata.Album = l.Album
	document.Metadata.DurationMS = l.DurationMS
	document.Metadata.Instrumental = l.Instrumental
	if l.Instrumental {
		return yaml.Marshal(document)
	}
	if len(l.Synced) > MaxLines {
		return nil, ErrTooLarge
	}
	wordCount := 0
	for _, line := range l.Synced {
		wordCount += len(line.Words)
		if wordCount > MaxWords {
			return nil, ErrTooLarge
		}
	}
	lines := slices.Clone(l.Synced)
	sortLines(lines)
	document.Lines = make([]lyricsfileLine, 0, len(lines))
	texts := make([]string, 0, len(lines))
	for _, line := range lines {
		start, err := stampMS(line.Start)
		if err != nil {
			return nil, err
		}
		text := line.Text
		if text == "" {
			text = joinWords(line.Words)
		}
		entry := lyricsfileLine{}
		entry.Text = text
		if line.End != 0 {
			end, err := stampMS(line.End)
			if err != nil || end < start {
				return nil, errLyricsfile
			}
			entry.EndMS = &end
		}
		if spans, keep := lyricsfileSpans(line.Words, text); keep {
			for _, span := range spans {
				start = min(start, *span.StartMS)
			}
			entry.Words = spans
		}
		entry.StartMS = &start
		texts = append(texts, text)
		document.Lines = append(document.Lines, entry)
	}
	plain := l.Plain
	if plain == "" {
		plain = strings.Join(texts, "\n")
	}
	if strings.Count(plain, "\n")+1 > MaxLines {
		return nil, ErrTooLarge
	}
	document.Plain = plain
	return yaml.Marshal(document)
}

// stampMS is milliseconds in reverse: Seconds to whole milliseconds for YAML.
func stampMS(seconds Seconds) (int64, error) {
	if !seconds.finite() || seconds < 0 || seconds > Seconds(math.MaxInt64/1000) {
		return 0, errLyricsfile
	}
	return int64(math.Round(float64(seconds) * 1000)), nil
}

// lyricsfileSpans converts words when they are complete, timed, and spell out
// line text: the same condition ParseLyricsfile uses to keep them.
func lyricsfileSpans(words []Word, lineText string) ([]lyricsfileSpan, bool) {
	if len(words) == 0 {
		return nil, false
	}
	joined := joinWords(words)
	if joined != lineText || strings.TrimSpace(joined) == "" {
		return nil, false
	}
	spans := make([]lyricsfileSpan, 0, len(words))
	for _, word := range words {
		start, err := stampMS(word.Start)
		if err != nil {
			return nil, false
		}
		span := lyricsfileSpan{Text: word.Text, StartMS: &start}
		if word.End != 0 {
			end, err := stampMS(word.End)
			if err != nil || end < start {
				return nil, false
			}
			span.EndMS = &end
		}
		spans = append(spans, span)
	}
	return spans, true
}

func validateLyricsfileNode(node *yaml.Node, depth int) error {
	if depth > 16 {
		return ErrTooLarge
	}
	if node.Anchor != "" || node.Kind == yaml.AliasNode {
		return errLyricsfile
	}
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str", "!!int", "!!bool", "!!null":
			return nil
		}
		return errLyricsfile
	case yaml.MappingNode:
		if node.Tag != "!!map" {
			return errLyricsfile
		}
		if len(node.Content) > 128 {
			return ErrTooLarge
		}
		keys := make(map[string]bool, len(node.Content)/2)
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || keys[key.Value] {
				return errLyricsfile
			}
			keys[key.Value] = true
		}
	case yaml.SequenceNode:
		if node.Tag != "!!seq" {
			return errLyricsfile
		}
	case yaml.DocumentNode:
	default:
		return errLyricsfile
	}
	for _, child := range node.Content {
		if err := validateLyricsfileNode(child, depth+1); err != nil {
			return err
		}
	}
	return nil
}
