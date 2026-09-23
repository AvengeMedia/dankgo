package files

import (
	"iter"
	"path/filepath"
	"slices"
	"strings"
)

func chunkEntries(entries []Entry, size int) iter.Seq[[]Entry] {
	return func(yield func([]Entry) bool) {
		for chunk := range slices.Chunk(entries, size) {
			if !yield(slices.Clone(chunk)) {
				return
			}
		}
	}
}

func splitPath(path string) (string, string) {
	return filepath.Dir(path), filepath.Base(path)
}

func isPath(value string) bool { return strings.HasPrefix(value, "/") }
