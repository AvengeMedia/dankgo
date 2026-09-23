package files

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AvengeMedia/dankgo/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func filterTree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "sub"), 0o755))
	for _, name := range files {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}
	return root
}

func entryNames(t *testing.T, result map[string]any) []string {
	t.Helper()
	raw, ok := result["entries"].([]any)
	require.True(t, ok)
	out := make([]string, len(raw))
	for i, e := range raw {
		out[i] = e.(map[string]any)["name"].(string)
	}
	return out
}

func TestHandleListFilters(t *testing.T) {
	root := filterTree(t, "a.PNG", "b.txt", "c.jpg", "notes.md")

	cases := []struct {
		name    string
		filters any
		want    []string
	}{
		{"no filters", nil, []string{"sub", "a.PNG", "b.txt", "c.jpg", "notes.md"}},
		{"empty list", []any{}, []string{"sub", "a.PNG", "b.txt", "c.jpg", "notes.md"}},
		{"case insensitive", []any{"*.png"}, []string{"sub", "a.PNG"}},
		{"upper case pattern", []any{"*.TXT"}, []string{"sub", "b.txt"}},
		{"any pattern matches", []any{"*.png", "*.jpg"}, []string{"sub", "a.PNG", "c.jpg"}},
		{"comma separated", "*.md,*.txt", []string{"sub", "b.txt", "notes.md"}},
		{"comma separated with spaces", "*.md, *.txt", []string{"sub", "b.txt", "notes.md"}},
		{"padded pattern", []any{" *.png "}, []string{"sub", "a.PNG"}},
		{"star dot star is a no-op", []any{"*.png", "*.*"}, []string{"sub", "a.PNG", "b.txt", "c.jpg", "notes.md"}},
		{"star is a no-op", []any{"*"}, []string{"sub", "a.PNG", "b.txt", "c.jpg", "notes.md"}},
		{"nothing matches keeps directories", []any{"*.pdf"}, []string{"sub"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestService(t)
			params := map[string]any{"path": root}
			if tc.filters != nil {
				params["filters"] = tc.filters
			}
			result := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.list", Params: params}))
			assert.Equal(t, tc.want, entryNames(t, result))
			assert.Equal(t, float64(len(tc.want)), result["total"])
		})
	}
}

func TestHandleInvalidFilterIsEINVAL(t *testing.T) {
	svc := newTestService(t)
	root := filterTree(t, "a.txt")

	for _, method := range []string{"files.list", "files.watch"} {
		out := call(t, svc, ipc.Request{ID: 1, Method: method, Params: map[string]any{"path": root, "filters": []any{"*.txt", "[a-"}}})
		assert.Equal(t, "EINVAL", out["code"], method)
		assert.NotContains(t, out, "result", method)
	}
	svc.mu.Lock()
	defer svc.mu.Unlock()
	assert.Empty(t, svc.watches, "a rejected watch must not be opened")
}

func TestWatchFiltersEvents(t *testing.T) {
	bus := &fakeBus{}
	svc := NewService(bus, t.TempDir(), nil)
	t.Cleanup(svc.Close)
	root := filterTree(t, "a.txt", "c.txt", "z.png")

	filters, err := parseFilters([]string{"*.txt"})
	require.NoError(t, err)
	w, err := svc.openWatch(context.Background(), root, ListOptions{Sort: nameSort(), Filters: filters})
	require.NoError(t, err)
	svc.Attach([]string{w.topic})

	require.NoError(t, os.WriteFile(filepath.Join(root, "b.png"), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "b.txt"), nil, 0o644))
	added := bus.await(t, func(b batch) bool { return len(b.Added) > 0 && b.Added[0].Name == "b.txt" })
	assert.Equal(t, []int{2}, added.AddedAt, "sub, a.txt, b.txt")

	require.NoError(t, os.Remove(filepath.Join(root, "z.png")))
	require.NoError(t, os.Remove(filepath.Join(root, "c.txt")))
	bus.await(t, func(b batch) bool { return len(b.Removed) > 0 && b.Removed[0] == "c.txt" })

	time.Sleep(2 * coalesceWindow)
	assertNeverMentioned(t, bus.snapshot(), "b.png", "z.png")
	assert.Equal(t, []string{"sub", "a.txt", "b.txt"}, viewNames(w, ListOptions{Sort: nameSort(), Filters: filters}))
}

func TestWatchPageChangesFilters(t *testing.T) {
	bus := &fakeBus{}
	svc := NewService(bus, t.TempDir(), nil)
	t.Cleanup(svc.Close)
	root := filterTree(t, "a.txt", "b.png")

	opened := resultOf(t, call(t, svc, ipc.Request{ID: 1, Method: "files.watch", Params: map[string]any{"path": root}}))
	assert.Equal(t, []string{"sub", "a.txt", "b.png"}, entryNames(t, opened))
	watchID := opened["watchId"].(string)
	svc.Attach([]string{"files:" + watchID})

	paged := resultOf(t, call(t, svc, ipc.Request{ID: 2, Method: "files.list", Params: map[string]any{"watchId": watchID, "filters": []any{"*.PNG"}}}))
	assert.Equal(t, []string{"sub", "b.png"}, entryNames(t, paged))
	assert.Equal(t, float64(2), paged["total"])

	require.NoError(t, os.WriteFile(filepath.Join(root, "c.txt"), nil, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "d.png"), nil, 0o644))
	added := bus.await(t, func(b batch) bool { return len(b.Added) > 0 && b.Added[0].Name == "d.png" })
	assert.Equal(t, []int{2}, added.AddedAt, "sub, b.png, d.png")

	time.Sleep(2 * coalesceWindow)
	assertNeverMentioned(t, bus.snapshot(), "c.txt")

	cleared := resultOf(t, call(t, svc, ipc.Request{ID: 3, Method: "files.list", Params: map[string]any{"watchId": watchID}}))
	assert.Equal(t, []string{"sub", "a.txt", "b.png", "c.txt", "d.png"}, entryNames(t, cleared))
}

func assertNeverMentioned(t *testing.T, events []batch, excluded ...string) {
	t.Helper()
	for _, event := range events {
		mentioned := append(namesOf(event.Added), namesOf(event.Changed)...)
		mentioned = append(mentioned, namesOf(event.Thumbnails)...)
		mentioned = append(mentioned, event.Removed...)
		for _, name := range excluded {
			assert.NotContains(t, mentioned, name, "event %+v", event)
		}
	}
}
