package files

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/AvengeMedia/dankgo/trash"
)

type trashFailure struct {
	Path  string `json:"path"`
	Code  Code   `json:"code"`
	Error string `json:"error"`
}

type trashResult struct {
	Trashed []string       `json:"trashed"`
	Failed  []trashFailure `json:"failed"`
}

func (s *Service) makeDir(path string) (Entry, error) {
	if !filepath.IsAbs(path) {
		return Entry{}, &Error{Code: CodeInvalid, Path: path, Err: errors.New("path must be absolute")}
	}
	path = filepath.Clean(path)
	if err := os.Mkdir(path, 0o755); err != nil {
		return Entry{}, Wrap(path, err)
	}
	return s.Stat(path)
}

func (s *Service) rename(path, name string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", &Error{Code: CodeInvalid, Path: path, Err: errors.New("path must be absolute")}
	}
	if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
		return "", &Error{Code: CodeInvalid, Path: name, Err: fmt.Errorf("invalid name %q", name)}
	}

	path = filepath.Clean(path)
	target := filepath.Join(filepath.Dir(path), name)
	if err := renameNoReplace(path, target); err != nil {
		return "", Wrap(path, err)
	}
	return target, nil
}

func renameIfAbsent(from, to string) error {
	_, err := os.Lstat(to)
	switch {
	case err == nil:
		return &os.LinkError{Op: "rename", Old: from, New: to, Err: fs.ErrExist}
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}
	return os.Rename(from, to)
}

func (s *Service) trashPaths(paths []string) trashResult {
	result := trashResult{Trashed: []string{}, Failed: []trashFailure{}}
	for _, path := range paths {
		if _, err := trash.Put(path); err != nil {
			failure := Wrap(path, err)
			result.Failed = append(result.Failed, trashFailure{Path: path, Code: CodeOf(failure), Error: failure.Error()})
			continue
		}
		result.Trashed = append(result.Trashed, path)
	}
	return result
}
