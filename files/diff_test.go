package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/AvengeMedia/dankgo/ipc"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func applyLikeClient(t *testing.T, list []string, b batch) []string {
	t.Helper()
	out := slices.Clone(list)
	for _, name := range b.Removed {
		if i := slices.Index(out, name); i >= 0 {
			out = slices.Delete(out, i, i+1)
		}
	}
	require.Len(t, b.AddedAt, len(b.Added))
	for i, e := range b.Added {
		require.LessOrEqual(t, b.AddedAt[i], len(out), "insert index past the end")
		out = slices.Insert(out, b.AddedAt[i], e.Name)
	}
	return out
}

func openFakeWatch(t *testing.T, root string, opts ListOptions) (*fakeBus, *fakeNotifier, *watch) {
	t.Helper()
	bus := &fakeBus{}
	svc := NewService(bus, t.TempDir(), nil)
	t.Cleanup(svc.Close)
	fake := newFakeNotifier()
	svc.newNotifier = func() (notifier, error) { return fake, nil }
	svc.AttachOnOpen()

	w, err := svc.openWatch(context.Background(), root, opts)
	require.NoError(t, err)
	svc.pool.drain()
	return bus, fake, w
}

func flushedBatch(t *testing.T, bus *fakeBus, seq int) batch {
	t.Helper()
	for _, event := range bus.snapshot() {
		if event.Seq == seq {
			return event
		}
	}
	t.Fatalf("no batch with seq %d in %+v", seq, bus.snapshot())
	return batch{}
}

func writeAged(t *testing.T, root, name string, age time.Duration, data string) {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.WriteFile(path, []byte(data), 0o644))
	stamp := time.Now().Add(-time.Hour + age)
	require.NoError(t, os.Chtimes(path, stamp, stamp))
}

func TestWatchBatchReplaysToTheServerView(t *testing.T) {
	txtOnly, err := parseFilters([]string{"*.txt"})
	require.NoError(t, err)
	pngOnly, err := parseFilters([]string{"*.png"})
	require.NoError(t, err)

	cases := []struct {
		name    string
		opts    ListOptions
		files   []string
		change  func(t *testing.T, root string)
		touched []string
	}{
		{
			name:  "delete and move in one window",
			opts:  ListOptions{Sort: SortSpec{Key: SortMtime, DirsFirst: true}},
			files: []string{"a", "b", "c", "d"},
			change: func(t *testing.T, root string) {
				require.NoError(t, os.Remove(filepath.Join(root, "a")))
				writeAged(t, root, "c", 10*time.Minute, "x")
			},
			touched: []string{"a", "c"},
		},
		{
			name:  "move forward past an untouched entry",
			opts:  ListOptions{Sort: SortSpec{Key: SortMtime, DirsFirst: true}},
			files: []string{"a", "b", "c", "d"},
			change: func(t *testing.T, root string) {
				writeAged(t, root, "d", -time.Minute, "x")
				writeAged(t, root, "e", 90*time.Second, "x")
			},
			touched: []string{"d", "e"},
		},
		{
			name:  "grow to the end under size sort",
			opts:  ListOptions{Sort: SortSpec{Key: SortSize, DirsFirst: true}},
			files: []string{"a", "b", "c"},
			change: func(t *testing.T, root string) {
				writeAged(t, root, "a", 0, "much bigger now")
			},
			touched: []string{"a"},
		},
		{
			name:  "rename out of the filter",
			opts:  ListOptions{Sort: nameSort(), Filters: txtOnly},
			files: []string{"a.txt", "b.png", "c.txt"},
			change: func(t *testing.T, root string) {
				require.NoError(t, os.Rename(filepath.Join(root, "a.txt"), filepath.Join(root, "a.png")))
			},
			touched: []string{"a.txt", "a.png"},
		},
		{
			name:  "rename into the filter",
			opts:  ListOptions{Sort: nameSort(), Filters: pngOnly},
			files: []string{"a.txt", "b.png", "c.txt"},
			change: func(t *testing.T, root string) {
				require.NoError(t, os.Rename(filepath.Join(root, "c.txt"), filepath.Join(root, "a.png")))
			},
			touched: []string{"c.txt", "a.png"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			for i, name := range tc.files {
				writeAged(t, root, name, time.Duration(i)*time.Minute, "x")
			}
			bus, _, w := openFakeWatch(t, root, tc.opts)
			page, seq := w.page(tc.opts)
			before := namesOf(page.Entries)

			tc.change(t, root)
			w.flush(tc.touched)

			replayed := applyLikeClient(t, before, flushedBatch(t, bus, seq+1))
			assert.Equal(t, viewNames(w, tc.opts), replayed)
		})
	}
}

func TestWatchInPlaceChangeStaysChanged(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.zip", "b.zip", "c.zip"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}
	opts := ListOptions{Sort: nameSort()}
	bus, _, w := openFakeWatch(t, root, opts)
	_, seq := w.page(opts)

	require.NoError(t, os.WriteFile(filepath.Join(root, "b.zip"), []byte("grown"), 0o644))
	w.flush([]string{"b.zip"})

	changes := flushedBatch(t, bus, seq+1)
	assert.Empty(t, changes.Removed)
	assert.Empty(t, changes.Added)
	require.Len(t, changes.Changed, 1)
	assert.Equal(t, "b.zip", changes.Changed[0].Name)
	assert.Equal(t, int64(5), changes.Changed[0].Size)
}

func TestWatchPageIsASnapshot(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.zip", "b.zip", "c.zip"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}
	opts := ListOptions{Sort: nameSort(), IncludeHidden: true}
	_, _, w := openFakeWatch(t, root, opts)
	page, _ := w.page(opts)

	require.NoError(t, os.Remove(filepath.Join(root, "a.zip")))
	w.flush([]string{"a.zip"})

	assert.Equal(t, []string{"a.zip", "b.zip", "c.zip"}, namesOf(page.Entries), "a later flush must not rewrite a page already handed out")
}

func TestWatchListingIsEncodedWhileFlushing(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a.zip", "b.zip", "c.zip", "d.zip"} {
		require.NoError(t, os.WriteFile(filepath.Join(root, name), nil, 0o644))
	}
	_, _, w := openFakeWatch(t, root, ListOptions{Sort: nameSort(), IncludeHidden: true})

	stop := make(chan struct{})
	flushed := make(chan struct{})
	go func() {
		defer close(flushed)
		churn(stop, w, filepath.Join(root, "a.zip"))
	}()
	for range 200 {
		call(t, w.svc, ipc.Request{ID: 1, Method: "files.list", Params: map[string]any{"watchId": w.id, "includeHidden": true}})
	}
	close(stop)
	<-flushed
}

func TestWatchListSeqMatchesItsEntries(t *testing.T) {
	root := t.TempDir()
	bus, _, w := openFakeWatch(t, root, ListOptions{Sort: nameSort()})

	type observed struct {
		seq     int
		present bool
	}
	stop := make(chan struct{})
	flushed := make(chan struct{})
	go func() {
		defer close(flushed)
		churn(stop, w, filepath.Join(root, "a.zip"))
	}()
	seen := make([]observed, 0, 500)
	for range 500 {
		result := resultOf(t, call(t, w.svc, ipc.Request{ID: 1, Method: "files.list", Params: map[string]any{"watchId": w.id}}))
		seen = append(seen, observed{seq: int(result["seq"].(float64)), present: slices.Contains(entryNames(t, result), "a.zip")})
	}
	close(stop)
	<-flushed

	present := map[int]bool{0: false}
	state := false
	for _, event := range bus.snapshot() {
		switch {
		case len(event.Added) > 0:
			state = true
		case len(event.Removed) > 0:
			state = false
		}
		present[event.Seq] = state
	}
	for _, o := range seen {
		require.Equal(t, present[o.seq], o.present, "the entries of a response must reflect every batch up to its seq %d", o.seq)
	}
}

func churn(stop <-chan struct{}, w *watch, path string) {
	for i := 0; ; i++ {
		select {
		case <-stop:
			return
		default:
		}
		switch i % 2 {
		case 0:
			_ = os.WriteFile(path, nil, 0o644)
		default:
			_ = os.Remove(path)
		}
		w.flush([]string{filepath.Base(path)})
	}
}

func TestWatchEnrichmentNeverReadsTheLiveListing(t *testing.T) {
	root := t.TempDir()
	for i := range 40 {
		require.NoError(t, os.WriteFile(filepath.Join(root, fmt.Sprintf("f%02d.txt", i)), []byte(fmt.Sprint(i)), 0o644))
	}
	bus := &fakeBus{}
	svc := NewService(bus, t.TempDir(), nil)
	t.Cleanup(svc.Close)
	fake := newFakeNotifier()
	svc.newNotifier = func() (notifier, error) { return fake, nil }
	svc.AttachOnOpen()

	resort := func(w *watch) {
		for _, key := range []SortKey{SortSize, SortMtime, SortName} {
			w.page(ListOptions{Sort: SortSpec{Key: key, Desc: true}})
		}
	}
	resorted := make(chan struct{})
	go func() {
		defer close(resorted)
		for {
			if w, ok := svc.watchByID("w1"); ok {
				resort(w)
				return
			}
			runtime.Gosched()
		}
	}()
	w, err := svc.openWatch(context.Background(), root, ListOptions{Sort: nameSort()})
	require.NoError(t, err)
	<-resorted

	fake.errors <- fsnotify.ErrEventOverflow
	require.Eventually(t, func() bool {
		resort(w)
		return slices.ContainsFunc(bus.snapshot(), func(b batch) bool { return b.Kind == "resync" })
	}, 5*time.Second, time.Millisecond)
	resort(w)
	svc.pool.drain()
}
