package files

import (
	"fmt"
	"path/filepath"
	"strings"
)

type nameFilter []string

func parseFilters(patterns []string) (nameFilter, error) {
	filter := make(nameFilter, 0, len(patterns))
	matchAll := false
	for _, raw := range patterns {
		pattern := strings.ToLower(strings.TrimSpace(raw))
		if _, err := filepath.Match(pattern, ""); err != nil {
			return nil, &Error{Code: CodeInvalid, Path: raw, Err: fmt.Errorf("invalid filter %q: %w", raw, err)}
		}
		switch pattern {
		case "":
			continue
		case "*", "*.*":
			matchAll = true
		}
		filter = append(filter, pattern)
	}
	if matchAll {
		return nil, nil
	}
	return filter, nil
}

func (f nameFilter) passes(e Entry) bool {
	if len(f) == 0 || e.IsDir {
		return true
	}
	name := strings.ToLower(e.Name)
	for _, pattern := range f {
		if matched, _ := filepath.Match(pattern, name); matched {
			return true
		}
	}
	return false
}

func filterView(entries []Entry, includeHidden bool, filters nameFilter) []Entry {
	if includeHidden && len(filters) == 0 {
		return entries
	}
	out := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if (e.Hidden && !includeHidden) || !filters.passes(e) {
			continue
		}
		out = append(out, e)
	}
	return out
}
