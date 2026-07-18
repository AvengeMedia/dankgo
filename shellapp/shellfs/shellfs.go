// Package shellfs materializes an embedded UI tree onto the filesystem,
// since quickshell needs real paths. Trees are extracted read-only, keyed
// by a build-time revision file (.dankrev) when the tree carries one, else
// by content hash computed and verified on every resolution.
package shellfs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	hashLen = 16
	// RevFile names a build-time revision key inside the UI tree (8-64 hex
	// chars). It spares every resolution a full content hash: the extracted
	// dir is trusted when its RevFile matches the embedded one.
	RevFile = ".dankrev"
)

func Extract(fsys fs.FS, baseDir string) (string, error) {
	if key, ok := embeddedRev(fsys); ok {
		return extractKeyed(fsys, baseDir, key)
	}

	hash, err := hashFS(fsys)
	if err != nil {
		return "", fmt.Errorf("hash embedded UI: %w", err)
	}

	target := filepath.Join(baseDir, hash)
	if verify(target, hash) {
		return target, nil
	}
	return materialize(fsys, baseDir, target, func() bool { return verify(target, hash) })
}

func extractKeyed(fsys fs.FS, baseDir, key string) (string, error) {
	target := filepath.Join(baseDir, key)
	if extractedRev(target) == key {
		return target, nil
	}
	return materialize(fsys, baseDir, target, func() bool { return extractedRev(target) == key })
}

func materialize(fsys fs.FS, baseDir, target string, extractedByOther func() bool) (string, error) {
	forceRemoveAll(target)

	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return "", err
	}

	tmp, err := os.MkdirTemp(baseDir, ".extract-")
	if err != nil {
		return "", err
	}
	defer forceRemoveAll(tmp)

	if err := os.CopyFS(tmp, fsys); err != nil {
		return "", fmt.Errorf("extract embedded UI: %w", err)
	}
	if err := makeReadOnly(tmp); err != nil {
		return "", fmt.Errorf("extract embedded UI: %w", err)
	}

	if err := os.Rename(tmp, target); err != nil {
		// A concurrent invocation may have extracted the same revision first.
		if extractedByOther() {
			return target, nil
		}
		return "", err
	}
	return target, nil
}

// Prune removes previously extracted revisions under baseDir, keeping
// keep. Callers must ensure no running instance still uses the others.
func Prune(baseDir, keep string) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		p := filepath.Join(baseDir, entry.Name())
		if !entry.IsDir() || p == keep {
			continue
		}
		forceRemoveAll(p)
	}
}

func embeddedRev(fsys fs.FS) (string, bool) {
	data, err := fs.ReadFile(fsys, RevFile)
	if err != nil {
		return "", false
	}
	key := strings.TrimSpace(string(data))
	if len(key) < 8 || len(key) > 64 {
		return "", false
	}
	for _, c := range key {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		default:
			return "", false
		}
	}
	return key, true
}

func extractedRev(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, RevFile))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func verify(dir, hash string) bool {
	if _, err := os.Stat(dir); err != nil {
		return false
	}
	got, err := hashFS(os.DirFS(dir))
	return err == nil && got == hash
}

func hashFS(fsys fs.FS) (string, error) {
	h := sha256.New()
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := fsys.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		h.Write([]byte(p))
		h.Write([]byte{0})
		_, err = io.Copy(h, f)
		return err
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil))[:hashLen], nil
}

func makeReadOnly(root string) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.Chmod(p, 0o555)
		}
		return os.Chmod(p, 0o444)
	})
}

// forceRemoveAll removes a read-only extraction, restoring dir write
// permission first so unlinking succeeds.
func forceRemoveAll(root string) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && d.IsDir() {
			_ = os.Chmod(p, 0o755)
		}
		return nil
	})
	os.RemoveAll(root)
}
