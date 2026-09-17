package lyrics

import (
	"errors"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

var errLyricsfile = errors.New("invalid Lyricsfile document")

type lyricsfileDocument struct {
	Version  string `yaml:"version"`
	Metadata struct {
		Instrumental bool `yaml:"instrumental"`
	} `yaml:"metadata"`
	Lines []lyricsfileLine `yaml:"lines"`
	Plain string           `yaml:"plain"`
}

type lyricsfileLine struct {
	lyricsfileSpan `yaml:",inline"`
	Words          []lyricsfileSpan `yaml:"words"`
}

type lyricsfileSpan struct {
	Text    string `yaml:"text"`
	StartMS *int64 `yaml:"start_ms"`
	EndMS   *int64 `yaml:"end_ms"`
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
		return &Lyrics{Instrumental: true}, nil
	}
	if len(document.Lines) > MaxLines || strings.Count(document.Plain, "\n")+1 > MaxLines {
		return nil, ErrTooLarge
	}
	result := &Lyrics{Plain: document.Plain}
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
