package files

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/AvengeMedia/dankgo/log"
)

type Publisher interface {
	Publish(topic string, data any)
}

type Service struct {
	bus     Publisher
	builder *builder
	pool    *pool
	thumbs  *Thumbs
	counts  *countCache

	newNotifier  func() (notifier, error)
	attachOnOpen atomic.Bool

	mu       sync.Mutex
	watches  map[string]*watch
	nextID   int
	scans    map[string]*scan
	nextScan int
}

func NewService(bus Publisher, cacheHome string, resolver MimeResolver) *Service {
	thumbs := NewThumbs(cacheHome)
	return &Service{
		bus:         bus,
		builder:     newBuilder(resolver, thumbs.Supports),
		pool:        newPool(min(4, runtime.NumCPU())),
		thumbs:      thumbs,
		counts:      newCountCache(),
		newNotifier: newFsnotifyWatcher,
		watches:     map[string]*watch{},
		scans:       map[string]*scan{},
	}
}

func (s *Service) Capabilities() []string { return s.thumbs.Capabilities() }

func (s *Service) Close() {
	s.mu.Lock()
	watches := make([]*watch, 0, len(s.watches))
	for _, w := range s.watches {
		watches = append(watches, w)
	}
	s.mu.Unlock()

	for _, w := range watches {
		w.close()
	}
	s.closeScans()
}

func (s *Service) register(w *watch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	w.id = "w" + strconv.Itoa(s.nextID)
	w.topic = "files:" + w.id
	s.watches[w.id] = w
}

func (s *Service) unregister(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.watches, id)
}

func (s *Service) watchByID(id string) (*watch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.watches[id]
	return w, ok
}

// Attach flushes what each watch buffered before its topic had a subscriber.
func (s *Service) Attach(topics []string) {
	for _, topic := range topics {
		id, ok := strings.CutPrefix(topic, "files:")
		if !ok {
			continue
		}
		if w, live := s.watchByID(id); live {
			w.attach()
		}
	}
}

// AttachOnOpen is for hosts without per-topic subscribe; later watches never buffer.
func (s *Service) AttachOnOpen() { s.attachOnOpen.Store(true) }

func (s *Service) publish(topic string, payload any) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(topic, payload)
}

func (s *Service) log(format string, args ...any) { log.Debugf(format, args...) }

func (s *Service) list(path string, opts ListOptions) (Page, error) {
	entries, err := s.builder.readDir(path)
	if err != nil {
		return Page{}, err
	}
	entries = filterView(entries, opts.IncludeHidden, opts.Filters)
	sortEntries(entries, opts.Sort)
	return pageOf(entries, opts.Cursor, opts.Limit), nil
}

func (s *Service) Stat(path string) (Entry, error) {
	dir, name := splitPath(path)
	entry, err := s.builder.entry(dir, name, hiddenNames(dir)[name])
	if err != nil {
		return Entry{}, Wrap(path, err)
	}
	if needsSniff(entry.Mime) && !entry.IsDir && !entry.SymlinkBroken {
		if sniffed := s.builder.mime.Sniff(entry.Path); sniffed != "" {
			entry.Mime = sniffed
		}
	}
	entry.Thumbnailable = s.thumbs.Supports(entry.Mime)
	if entry.Thumbnailable {
		entry.Thumbnail, _ = s.thumbs.Lookup(entry.Path, entry.MtimeMs, "normal")
	}
	entry.IconName = s.builder.icons.best(entry)
	s.builder.display(&entry)
	return entry, nil
}

func (s *Service) count(path string, includeHidden bool) Count {
	entry, err := s.Stat(path)
	if err != nil {
		return Count{Code: string(CodeOf(err))}
	}
	if !entry.IsDir {
		return Count{Code: string(CodeNotDir)}
	}
	if cached, ok := s.counts.get(path, entry.MtimeMs); ok {
		return cached
	}
	counted := countChildren(path, includeHidden)
	s.counts.put(path, entry.MtimeMs, counted)
	return counted
}

type thumbResult struct {
	Path      string `json:"path"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Failed    bool   `json:"failed,omitempty"`
	Pending   bool   `json:"pending,omitempty"`
}

func (s *Service) thumbnails(paths []string, size string, w *watch) []thumbResult {
	results := make([]thumbResult, 0, len(paths))
	queue := make([]Entry, 0, len(paths))

	for _, path := range paths {
		entry, ok := s.thumbEntry(w, path)
		switch {
		case !ok:
			results = append(results, thumbResult{Path: path, Failed: true})
		case !s.thumbs.Supports(entry.Mime):
			results = append(results, thumbResult{Path: path, Failed: true})
		default:
			cached, hit := s.thumbs.Lookup(path, entry.MtimeMs, size)
			switch {
			case hit:
				results = append(results, thumbResult{Path: path, Thumbnail: cached})
			case s.thumbs.Failed(path, entry.MtimeMs):
				results = append(results, thumbResult{Path: path, Failed: true})
			default:
				results = append(results, thumbResult{Path: path, Pending: true})
				queue = append(queue, entry)
			}
		}
	}

	switch w {
	case nil:
		return append(results[:0], s.generateNow(queue, size, results)...)
	default:
		s.queueThumbnails(queue, size, w)
	}
	return results
}

func (s *Service) generateNow(queue []Entry, size string, results []thumbResult) []thumbResult {
	generated := map[string]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, entry := range queue {
		wg.Add(1)
		s.pool.submit(func() {
			defer wg.Done()
			path, err := s.thumbs.Generate(entry, size)
			if err != nil {
				return
			}
			mu.Lock()
			generated[entry.Path] = path
			mu.Unlock()
		})
	}
	wg.Wait()

	for i, result := range results {
		path, ok := generated[result.Path]
		if !ok {
			results[i].Failed = result.Pending
			results[i].Pending = false
			continue
		}
		results[i] = thumbResult{Path: result.Path, Thumbnail: path}
	}
	return results
}

func (s *Service) queueThumbnails(queue []Entry, size string, w *watch) {
	for chunk := range chunkEntries(queue, thumbChunk) {
		work := chunk
		s.pool.submit(func() {
			done := make([]Entry, 0, len(work))
			for _, entry := range work {
				if w.closed() {
					return
				}
				path, err := s.thumbs.Generate(entry, size)
				if err != nil {
					continue
				}
				entry.Thumbnail = path
				done = append(done, entry)
			}
			w.applyChanged(done, "thumbnails")
		})
	}
}

func (s *Service) thumbEntry(w *watch, path string) (Entry, bool) {
	if w != nil {
		_, name := splitPath(path)
		if entry, ok := w.entry(name); ok && entry.Path == path {
			return entry, true
		}
	}
	entry, err := s.Stat(path)
	if err != nil {
		return Entry{}, false
	}
	return entry, true
}

func (s *Service) icons(keys []string) map[string][]string {
	out := make(map[string][]string, len(keys))
	for _, key := range keys {
		if mediaType(key) != "" && !isPath(key) {
			out[key] = mimeLadder(key, false)
			continue
		}
		entry, err := s.Stat(key)
		if err != nil {
			out[key] = []string{"emblem-unreadable", genericIcon}
			continue
		}
		out[key] = s.builder.icons.ladder(entry)
	}
	return out
}

func (s *Service) Watch(ctx context.Context, path string, spec SortSpec, includeHidden bool) (*watch, error) {
	return s.openWatch(ctx, path, ListOptions{Sort: spec, IncludeHidden: includeHidden})
}
