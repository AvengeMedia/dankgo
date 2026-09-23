package files

import (
	"io"
	"os"
	"strings"
	"sync"
)

const CountCap = 10000

type Count struct {
	Count  int    `json:"count"`
	Capped bool   `json:"capped"`
	Code   string `json:"code,omitempty"`
}

type countCache struct {
	mu      sync.Mutex
	entries map[string]countEntry
}

type countEntry struct {
	mtimeMs int64
	count   Count
}

func newCountCache() *countCache { return &countCache{entries: map[string]countEntry{}} }

func (c *countCache) get(path string, mtimeMs int64) (Count, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cached, ok := c.entries[path]
	if !ok || cached.mtimeMs != mtimeMs {
		return Count{}, false
	}
	return cached.count, true
}

func (c *countCache) put(path string, mtimeMs int64, count Count) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= 512 {
		clear(c.entries)
	}
	c.entries[path] = countEntry{mtimeMs: mtimeMs, count: count}
}

func countChildren(path string, includeHidden bool) Count {
	f, err := os.Open(path)
	if err != nil {
		return Count{Code: string(CodeOf(Wrap(path, err)))}
	}
	defer f.Close()

	total := 0
	for total <= CountCap {
		batch, err := f.ReadDir(256)
		for _, de := range batch {
			if !includeHidden && strings.HasPrefix(de.Name(), ".") {
				continue
			}
			total++
		}
		if err == io.EOF {
			return Count{Count: total}
		}
		if err != nil {
			return Count{Count: total, Code: string(CodeOf(Wrap(path, err)))}
		}
	}
	return Count{Count: CountCap, Capped: true}
}
