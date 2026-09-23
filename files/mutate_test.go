package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/AvengeMedia/dankgo/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func errorCode(t *testing.T, out map[string]any) string {
	t.Helper()
	assert.NotContains(t, out, "result")
	assert.NotEmpty(t, out["error"])
	code, _ := out["code"].(string)
	return code
}

func lockedDir(t *testing.T) string {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root bypasses permission bits")
	}
	dir := filepath.Join(t.TempDir(), "locked")
	require.NoError(t, os.Mkdir(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "inside.txt"), nil, 0o644))
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	return dir
}

func TestHandleMkdir(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	path := filepath.Join(root, "New Folder")

	result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.mkdir", Params: map[string]any{"path": path + "/"}}))
	assert.Equal(t, path, result["path"])
	entry := result["entry"].(map[string]any)
	assert.Equal(t, "New Folder", entry["name"])
	assert.Equal(t, path, entry["path"])
	assert.Equal(t, true, entry["isDir"])

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestHandleMkdirErrors(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "file"), nil, 0o644))

	cases := []struct {
		name string
		path string
		want string
	}{
		{"relative", "relative/dir", "EINVAL"},
		{"missing path", "", "EINVAL"},
		{"exists", filepath.Join(root, "file"), "EEXIST"},
		{"missing parent", filepath.Join(root, "nope", "child"), "ENOENT"},
		{"parent is a file", filepath.Join(root, "file", "child"), "ENOTDIR"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := call(t, svc, ipc.Request{ID: 1, Method: "files.mkdir", Params: map[string]any{"path": tc.path}})
			assert.Equal(t, tc.want, errorCode(t, out))
		})
	}

	out := call(t, svc, ipc.Request{ID: 2, Method: "files.mkdir", Params: map[string]any{"path": filepath.Join(lockedDir(t), "child")}})
	assert.Equal(t, "EACCES", errorCode(t, out))
}

func TestHandleRename(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "old.txt"), []byte("data"), 0o644))

	result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.rename", Params: map[string]any{"path": filepath.Join(root, "old.txt"), "name": "new name.txt"}}))
	assert.Equal(t, filepath.Join(root, "new name.txt"), result["path"])

	data, err := os.ReadFile(filepath.Join(root, "new name.txt"))
	require.NoError(t, err)
	assert.Equal(t, "data", string(data))
	assert.NoFileExists(t, filepath.Join(root, "old.txt"))
}

func TestHandleRenameErrors(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	source := filepath.Join(root, "a.txt")
	require.NoError(t, os.WriteFile(source, []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), []byte("b"), 0o644))

	cases := []struct {
		name   string
		params map[string]any
		want   string
	}{
		{"target exists", map[string]any{"path": source, "name": "b.txt"}, "EEXIST"},
		{"empty name", map[string]any{"path": source, "name": ""}, "EINVAL"},
		{"missing name", map[string]any{"path": source}, "EINVAL"},
		{"dot", map[string]any{"path": source, "name": "."}, "EINVAL"},
		{"dot dot", map[string]any{"path": source, "name": ".."}, "EINVAL"},
		{"slash", map[string]any{"path": source, "name": "sub/c.txt"}, "EINVAL"},
		{"relative path", map[string]any{"path": "a.txt", "name": "c.txt"}, "EINVAL"},
		{"missing source", map[string]any{"path": filepath.Join(root, "gone.txt"), "name": "c.txt"}, "ENOENT"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := call(t, svc, ipc.Request{ID: 1, Method: "files.rename", Params: tc.params})
			assert.Equal(t, tc.want, errorCode(t, out))
		})
	}

	a, err := os.ReadFile(source)
	require.NoError(t, err)
	b, err := os.ReadFile(filepath.Join(root, "b.txt"))
	require.NoError(t, err)
	assert.Equal(t, "a", string(a))
	assert.Equal(t, "b", string(b), "an existing target is never overwritten")

	locked := lockedDir(t)
	out := call(t, svc, ipc.Request{ID: 2, Method: "files.rename", Params: map[string]any{"path": filepath.Join(locked, "inside.txt"), "name": "moved.txt"}})
	assert.Equal(t, "EACCES", errorCode(t, out))
}

func TestHandleTrash(t *testing.T) {
	base := t.TempDir()
	dataHome := filepath.Join(base, "data")
	t.Setenv("XDG_DATA_HOME", dataHome)
	t.Setenv("HOME", base)

	svc := newTestService(t)
	root := filepath.Join(base, "work")
	require.NoError(t, os.Mkdir(root, 0o755))
	file := filepath.Join(root, "a.txt")
	dir := filepath.Join(root, "folder")
	missing := filepath.Join(root, "missing.txt")
	require.NoError(t, os.WriteFile(file, []byte("a"), 0o644))
	require.NoError(t, os.Mkdir(dir, 0o755))

	result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.trash", Params: map[string]any{"paths": []any{file, missing, dir}}}))
	assert.Equal(t, []any{file, dir}, result["trashed"])

	failed := result["failed"].([]any)
	require.Len(t, failed, 1)
	failure := failed[0].(map[string]any)
	assert.Equal(t, missing, failure["path"])
	assert.Equal(t, "ENOENT", failure["code"])
	assert.NotEmpty(t, failure["error"])

	assert.NoFileExists(t, file)
	assert.NoDirExists(t, dir)
	assert.FileExists(t, filepath.Join(dataHome, "Trash", "files", "a.txt"))
	assert.FileExists(t, filepath.Join(dataHome, "Trash", "info", "a.txt.trashinfo"))
	assert.DirExists(t, filepath.Join(dataHome, "Trash", "files", "folder"))
}

func TestHandleTrashEmptyResultIsArrays(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))
	svc := newTestService(t)

	result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.trash", Params: map[string]any{"paths": []any{filepath.Join(t.TempDir(), "gone")}}}))
	assert.Equal(t, []any{}, result["trashed"])
	assert.Len(t, result["failed"], 1)
}

func TestHandleTrashInvalidParams(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.Join(t.TempDir(), "data"))
	svc := newTestService(t)
	kept := filepath.Join(t.TempDir(), "kept.txt")
	require.NoError(t, os.WriteFile(kept, nil, 0o644))

	assert.Equal(t, "EINVAL", errorCode(t, call(t, svc, ipc.Request{ID: 1, Method: "files.trash"})))
	assert.Equal(t, "EINVAL", errorCode(t, call(t, svc, ipc.Request{ID: 2, Method: "files.trash", Params: map[string]any{"paths": []any{kept, "relative.txt"}}})))
	assert.FileExists(t, kept, "a request with a relative path trashes nothing")
}

func TestRenameUsesThePathAsSent(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("plain"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt "), []byte("trailing"), 0o644))

	resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.rename", Params: map[string]any{"path": filepath.Join(root, "a.txt "), "name": "b.txt"}}))

	renamed, err := os.ReadFile(filepath.Join(root, "b.txt"))
	require.NoError(t, err)
	assert.Equal(t, "trailing", string(renamed))
	kept, err := os.ReadFile(filepath.Join(root, "a.txt"))
	require.NoError(t, err)
	assert.Equal(t, "plain", string(kept))
}

func TestMkdirUsesThePathAsSent(t *testing.T) {
	svc := newTestService(t)
	path := filepath.Join(t.TempDir(), "New Folder ")

	result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.mkdir", Params: map[string]any{"path": path}}))
	assert.Equal(t, path, result["path"])
	assert.DirExists(t, path)
	assert.NoDirExists(t, filepath.Join(filepath.Dir(path), "New Folder"))
}

func TestTrashUsesThePathsAsSent(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(base, "data"))
	t.Setenv("HOME", base)
	svc := newTestService(t)
	root := filepath.Join(base, "work")
	require.NoError(t, os.Mkdir(root, 0o755))
	for _, name := range []string{"keep.txt", "keep.txt ", "x", "x,y.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}

	cases := []struct {
		name   string
		params map[string]any
		gone   string
	}{
		{"trailing space", map[string]any{"path": filepath.Join(root, "keep.txt ")}, "keep.txt "},
		{"comma in a single paths string", map[string]any{"paths": filepath.Join(root, "x,y.txt")}, "x,y.txt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.trash", Params: tc.params}))
			assert.Equal(t, []any{filepath.Join(root, tc.gone)}, result["trashed"])
			assert.NoFileExists(t, filepath.Join(root, tc.gone))
		})
	}
	assert.FileExists(t, filepath.Join(root, "keep.txt"))
	assert.FileExists(t, filepath.Join(root, "x"))
}

func TestStatUsesThePathAsSent(t *testing.T) {
	svc := newTestService(t)
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt"), []byte("1"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "a.txt "), []byte("12345"), 0o644))

	entry := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.stat", Params: map[string]any{"path": filepath.Join(root, "a.txt ")}}))["entry"].(map[string]any)
	assert.Equal(t, "a.txt ", entry["name"])
	assert.Equal(t, float64(5), entry["size"])
}

func TestRenameNeverReplacesAnExistingTarget(t *testing.T) {
	renames := map[string]func(from, to string) error{
		"renameNoReplace": renameNoReplace,
		"renameIfAbsent":  renameIfAbsent,
	}
	for name, rename := range renames {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			from, taken, free := filepath.Join(root, "a"), filepath.Join(root, "b"), filepath.Join(root, "c")
			require.NoError(t, os.WriteFile(from, []byte("a"), 0o644))
			require.NoError(t, os.WriteFile(taken, []byte("b"), 0o644))

			err := rename(from, taken)
			require.ErrorIs(t, err, fs.ErrExist)
			assert.Equal(t, CodeExists, CodeFromOS(err))
			data, readErr := os.ReadFile(taken)
			require.NoError(t, readErr)
			assert.Equal(t, "b", string(data))

			require.NoError(t, rename(from, free))
			data, readErr = os.ReadFile(free)
			require.NoError(t, readErr)
			assert.Equal(t, "a", string(data))
			assert.NoFileExists(t, from)
		})
	}
}
