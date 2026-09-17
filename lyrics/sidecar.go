package lyrics

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Tried in this order, first one with lyrics wins. A new format is a Parser plus a row here.
var sidecarFormats = []struct {
	suffix string
	parse  Parser
}{
	{".lyricsfile.yaml", ParseLyricsfile},
	{".ttml", ParseTTML},
	{".lrc", ParseLRC},
	{".LRC", ParseLRC},
}

func fromSidecar(fileURL string) (*Result, bool) {
	path, ok := localPath(fileURL)
	if !ok {
		return nil, false
	}

	base := strings.TrimSuffix(path, filepath.Ext(path))
	for _, format := range sidecarFormats {
		data, ok := readSidecar(base + format.suffix)
		if !ok {
			continue
		}
		parsed, err := format.parse(data)
		if err != nil || parsed.Empty() {
			continue
		}
		return &Result{Found: true, Source: Sidecar, Lyrics: *parsed}, true
	}
	return nil, false
}

func localPath(fileURL string) (string, bool) {
	if !strings.HasPrefix(fileURL, "file://") {
		return "", false
	}
	parsed, err := url.Parse(fileURL)
	if err != nil || (parsed.Host != "" && parsed.Host != "localhost") || !filepath.IsAbs(parsed.Path) {
		return "", false
	}
	return parsed.Path, true
}

func readSidecar(path string) ([]byte, bool) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > MaxSidecarBytes {
		return nil, false
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaxSidecarBytes+1))
	if err != nil || len(data) > MaxSidecarBytes {
		return nil, false
	}
	return data, true
}
