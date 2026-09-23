package files

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func nameSort() SortSpec { return SortSpec{Key: SortName, DirsFirst: true} }

func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "zeta"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(root, "Alpha"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(root, ".hiddendir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), []byte("12345"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "A.txt"), []byte("1"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".hidden"), nil, 0o644))
	require.NoError(t, os.Symlink(filepath.Join(root, "b.txt"), filepath.Join(root, "link")))
	require.NoError(t, os.Symlink(filepath.Join(root, "zeta"), filepath.Join(root, "dirlink")))
	require.NoError(t, os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "dangling")))
	return root
}

func byName(t *testing.T, entries []Entry, name string) Entry {
	t.Helper()
	for _, e := range entries {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("entry %q missing", name)
	return Entry{}
}

func TestListSortsDirectoriesFirst(t *testing.T) {
	root := buildTree(t)
	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	assert.Equal(t, []string{"Alpha", "dirlink", "zeta", "A.txt", "b.txt", "dangling", "link"}, namesOf(entries))
}

func TestListIncludesHiddenOnRequest(t *testing.T) {
	root := buildTree(t)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".hidden"), []byte("b.txt\n\n# not a comment\n"), 0o644))

	entries, err := List(root, ListOptions{IncludeHidden: true, Sort: nameSort()})
	require.NoError(t, err)
	assert.Contains(t, namesOf(entries), ".hiddendir")
	assert.True(t, byName(t, entries, "b.txt").Hidden, "names listed in .hidden are hidden")
	assert.False(t, byName(t, entries, "A.txt").Hidden)

	visible, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	assert.NotContains(t, namesOf(visible), "b.txt")
	assert.NotContains(t, namesOf(visible), ".hiddendir")
}

func TestListEntryFields(t *testing.T) {
	root := buildTree(t)
	stamp := time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC)
	require.NoError(t, os.Chtimes(filepath.Join(root, "b.txt"), stamp, stamp))

	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)

	file := byName(t, entries, "b.txt")
	assert.Equal(t, filepath.Join(root, "b.txt"), file.Path)
	assert.False(t, file.IsDir)
	assert.Equal(t, int64(5), file.Size)
	assert.Equal(t, stamp.UnixMilli(), file.MtimeMs)
	assert.Equal(t, "-rw-r--r--", file.Mode)
	assert.Equal(t, "txt", file.Extension)
	assert.Equal(t, "text/plain", file.Mime)
	assert.Equal(t, "text-plain", file.IconName)
	assert.False(t, file.Executable)
	assert.NotEmpty(t, file.Owner)
	assert.NotZero(t, file.CtimeMs)

	dir := byName(t, entries, "zeta")
	assert.True(t, dir.IsDir)
	assert.Equal(t, int64(-1), dir.Size, "directories carry -1 until counted")
	assert.Equal(t, MimeDirectory, dir.Mime)
	assert.Equal(t, "folder", dir.IconName)
}

func TestListSymlinks(t *testing.T) {
	root := buildTree(t)
	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)

	link := byName(t, entries, "link")
	assert.True(t, link.IsSymlink)
	assert.False(t, link.IsDir)
	assert.Equal(t, filepath.Join(root, "b.txt"), link.SymlinkTarget)
	assert.Equal(t, int64(5), link.Size)
	assert.False(t, link.SymlinkBroken)

	assert.True(t, byName(t, entries, "dirlink").IsDir)

	dangling := byName(t, entries, "dangling")
	assert.True(t, dangling.SymlinkBroken)
	assert.Equal(t, MimeSymlink, dangling.Mime)
	assert.Equal(t, "inode-symlink", dangling.IconName)
}

func TestListSymlinkLoopDoesNotHang(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.Symlink(filepath.Join(root, "b"), filepath.Join(root, "a")))
	require.NoError(t, os.Symlink(filepath.Join(root, "a"), filepath.Join(root, "b")))

	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, namesOf(entries))
	assert.True(t, byName(t, entries, "a").SymlinkBroken)
}

func TestListUnreadableChild(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits")
	}
	root := t.TempDir()
	locked := filepath.Join(root, "locked")
	require.NoError(t, os.Mkdir(locked, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(locked, "inner"), nil, 0o644))
	require.NoError(t, os.Chmod(locked, 0o000))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	assert.Equal(t, []string{"locked"}, namesOf(entries))

	_, err = List(locked, ListOptions{Sort: nameSort()})
	assert.Equal(t, CodeAccess, CodeOf(err))
}

func TestListLargeDirectory(t *testing.T) {
	root := t.TempDir()
	for i := range 3000 {
		require.NoError(t, os.WriteFile(filepath.Join(root, fmt.Sprintf("file-%d.txt", i)), nil, 0o644))
	}
	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	require.Len(t, entries, 3000)
	assert.Equal(t, []string{"file-0.txt", "file-1.txt", "file-2.txt"}, namesOf(entries[:3]))
	assert.Equal(t, "file-2999.txt", entries[2999].Name)
}

func TestListErrorCodes(t *testing.T) {
	root := buildTree(t)

	_, err := List("relative/path", ListOptions{})
	assert.Equal(t, CodeInvalid, CodeOf(err))

	_, err = List(filepath.Join(root, "nope"), ListOptions{})
	assert.Equal(t, CodeNotFound, CodeOf(err))

	_, err = List(filepath.Join(root, "b.txt"), ListOptions{})
	assert.Equal(t, CodeNotDir, CodeOf(err))
}

func TestListEmptyDirectory(t *testing.T) {
	entries, err := List(t.TempDir(), ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, entries)
	assert.NotNil(t, entries)
}

func TestPagingWalksEveryEntryOnce(t *testing.T) {
	root := t.TempDir()
	for i := range 12 {
		require.NoError(t, os.WriteFile(filepath.Join(root, fmt.Sprintf("f%02d", i)), nil, 0o644))
	}
	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)

	seen := []string{}
	cursor := ""
	for {
		page := pageOf(entries, cursor, 5)
		seen = append(seen, namesOf(page.Entries)...)
		assert.Equal(t, 12, page.Total)
		if page.Cursor == "" {
			break
		}
		cursor = page.Cursor
	}
	assert.Equal(t, namesOf(entries), seen)
}

func TestPagingCursorSurvivesInsertions(t *testing.T) {
	root := t.TempDir()
	for i := range 6 {
		require.NoError(t, os.WriteFile(filepath.Join(root, fmt.Sprintf("f%02d", i)), nil, 0o644))
	}
	entries, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)

	first := pageOf(entries, "", 3)
	assert.Equal(t, []string{"f00", "f01", "f02"}, namesOf(first.Entries))

	require.NoError(t, os.WriteFile(filepath.Join(root, "aaa"), nil, 0o644))
	grown, err := List(root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)

	second := pageOf(grown, first.Cursor, 3)
	assert.Equal(t, []string{"f03", "f04", "f05"}, namesOf(second.Entries), "an insert before the cursor must not skip entries")
	assert.Equal(t, 7, second.Total)
}

func TestPagingCursorFallsBackWhenEntryVanishes(t *testing.T) {
	entries := make([]Entry, 0, 5)
	for i := range 5 {
		entries = append(entries, Entry{Name: fmt.Sprintf("f%d", i)})
	}
	page := pageOf(entries[:4], "1:f1", 2)
	assert.Equal(t, []string{"f2", "f3"}, namesOf(page.Entries))

	gone := pageOf([]Entry{{Name: "f0"}, {Name: "f2"}, {Name: "f3"}}, "1:f1", 2)
	assert.Equal(t, []string{"f2", "f3"}, namesOf(gone.Entries))
}
