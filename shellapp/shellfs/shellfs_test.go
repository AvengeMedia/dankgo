package shellfs

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"shell.qml":            &fstest.MapFile{Data: []byte("import QtQuick\n")},
		"Common/Theme.qml":     &fstest.MapFile{Data: []byte("// theme\n")},
		"translations/en.json": &fstest.MapFile{Data: []byte("{}\n")},
	}
}

// tempBase returns a temp dir whose read-only extractions are made
// writable again before t.TempDir's own cleanup.
func tempBase(t *testing.T) string {
	base := t.TempDir()
	t.Cleanup(func() {
		filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() {
				os.Chmod(p, 0o755)
			}
			return nil
		})
	})
	return base
}

func keyedFS(key string) fstest.MapFS {
	m := testFS()
	m[RevFile] = &fstest.MapFile{Data: []byte(key + "\n")}
	return m
}

func TestExtractKeyedUsesRevAsDirName(t *testing.T) {
	base := tempBase(t)

	dir, err := Extract(keyedFS("abcdef0123456789"), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if filepath.Base(dir) != "abcdef0123456789" {
		t.Fatalf("dir %q, want rev-keyed name", dir)
	}
	if _, err := os.Stat(filepath.Join(dir, "shell.qml")); err != nil {
		t.Fatalf("shell.qml missing after extract: %v", err)
	}

	again, err := Extract(keyedFS("abcdef0123456789"), base)
	if err != nil {
		t.Fatalf("re-extract: %v", err)
	}
	if again != dir {
		t.Fatalf("re-extract returned %q, want %q", again, dir)
	}
}

func TestExtractKeyedNewRevExtractsFresh(t *testing.T) {
	base := tempBase(t)

	first, err := Extract(keyedFS("aaaaaaaaaaaaaaaa"), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	second, err := Extract(keyedFS("bbbbbbbbbbbbbbbb"), base)
	if err != nil {
		t.Fatalf("extract new rev: %v", err)
	}
	if second == first {
		t.Fatal("expected a new dir for a new revision")
	}
}

func TestExtractInvalidRevFallsBackToHashing(t *testing.T) {
	base := tempBase(t)

	dir, err := Extract(keyedFS("../../escape"), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if filepath.Dir(dir) != base {
		t.Fatalf("dir %q escaped base %q", dir, base)
	}
	if len(filepath.Base(dir)) != hashLen {
		t.Fatalf("dir %q, want content-hash name", dir)
	}
}

func TestExtractIsIdempotent(t *testing.T) {
	base := tempBase(t)

	first, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(first, "shell.qml")); err != nil {
		t.Fatalf("shell.qml missing after extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(first, "translations", "en.json")); err != nil {
		t.Fatalf("nested file missing after extract: %v", err)
	}

	second, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("re-extract: %v", err)
	}
	if second != first {
		t.Fatalf("re-extract returned %q, want %q", second, first)
	}
}

func TestExtractIsReadOnly(t *testing.T) {
	base := tempBase(t)

	dir, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "shell.qml"), []byte("hacked"), 0o644); err == nil {
		t.Fatal("expected write to extracted file to fail")
	}
	if err := os.WriteFile(filepath.Join(dir, "new.qml"), []byte("x"), 0o644); err == nil {
		t.Fatal("expected file creation in extracted dir to fail")
	}
}

func TestExtractHealsTamperedTree(t *testing.T) {
	base := tempBase(t)

	dir, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	tampered := filepath.Join(dir, "shell.qml")
	if err := os.Chmod(tampered, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tampered, []byte("hacked\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	healed, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("re-extract after tamper: %v", err)
	}
	if healed != dir {
		t.Fatalf("healed dir %q, want %q", healed, dir)
	}

	data, err := os.ReadFile(tampered)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "import QtQuick\n" {
		t.Fatalf("tampered file not restored, got %q", data)
	}
}

func TestExtractChangesDirWithContent(t *testing.T) {
	base := tempBase(t)

	first, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	changed := testFS()
	changed["shell.qml"] = &fstest.MapFile{Data: []byte("import QtQuick // v2\n")}
	second, err := Extract(changed, base)
	if err != nil {
		t.Fatalf("extract changed: %v", err)
	}
	if second == first {
		t.Fatal("expected a new dir for changed content")
	}
}

func TestPruneKeepsCurrent(t *testing.T) {
	base := tempBase(t)

	old, err := Extract(testFS(), base)
	if err != nil {
		t.Fatalf("extract old: %v", err)
	}

	changed := testFS()
	changed["shell.qml"] = &fstest.MapFile{Data: []byte("import QtQuick // v2\n")}
	current, err := Extract(changed, base)
	if err != nil {
		t.Fatalf("extract current: %v", err)
	}

	Prune(base, current)

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("stale read-only dir not pruned: %v", err)
	}
	if _, err := os.Stat(filepath.Join(current, "shell.qml")); err != nil {
		t.Fatalf("current dir pruned: %v", err)
	}
}
