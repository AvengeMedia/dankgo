package files

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	sizeTopic    = "sizes"
	sizeInterval = 500 * time.Millisecond
	sizeCheckRun = 256
)

// SizeEvent bytes are apparent sizes: symlinks are not followed, hardlinks count once per link.
type SizeEvent struct {
	ScanID    string `json:"scanId"`
	Bytes     int64  `json:"bytes"`
	Files     int    `json:"files"`
	Dirs      int    `json:"dirs"`
	Done      bool   `json:"done"`
	Cancelled bool   `json:"cancelled"`
	Code      string `json:"code,omitempty"`
}

type scan struct {
	id     string
	svc    *Service
	cancel context.CancelFunc
}

func (s *Service) startSize(ctx context.Context, paths []string) *scan {
	scanCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	sc := &scan{svc: s, cancel: cancel}
	sc.id = s.registerScan(sc)
	stopFromCtx := context.AfterFunc(ctx, cancel)
	go func() {
		defer stopFromCtx()
		sc.run(scanCtx, paths)
	}()
	return sc
}

func (s *Service) registerScan(sc *scan) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextScan++
	id := "s" + strconv.Itoa(s.nextScan)
	s.scans[id] = sc
	return id
}

func (s *Service) cancelScan(id string) bool {
	s.mu.Lock()
	sc, ok := s.scans[id]
	s.mu.Unlock()
	if !ok {
		return false
	}
	sc.close()
	return true
}

func (s *Service) closeScans() {
	s.mu.Lock()
	live := make([]*scan, 0, len(s.scans))
	for _, sc := range s.scans {
		live = append(live, sc)
	}
	s.mu.Unlock()
	for _, sc := range live {
		sc.close()
	}
}

func (sc *scan) close() { sc.cancel() }

func (sc *scan) run(ctx context.Context, paths []string) {
	event := SizeEvent{ScanID: sc.id}
	next := time.Now().Add(sizeInterval)
	seen := 0

	for _, path := range paths {
		if code := sc.walk(ctx, path, &event, &next, &seen); code != "" {
			event.Code = code
		}
		if ctx.Err() != nil {
			break
		}
	}

	event.Done = true
	event.Cancelled = ctx.Err() != nil
	sc.svc.unregisterScan(sc.id)
	sc.svc.publish(sizeTopic, event)
	sc.close()
}

func (sc *scan) walk(ctx context.Context, path string, event *SizeEvent, next *time.Time, seen *int) string {
	info, err := os.Lstat(path)
	if err != nil {
		return string(CodeOf(Wrap(path, err)))
	}
	if !info.IsDir() {
		event.Files++
		event.Bytes += info.Size()
		return ""
	}

	code := ""
	walkErr := filepath.WalkDir(path, func(child string, entry fs.DirEntry, err error) error {
		*seen++
		if *seen%sizeCheckRun == 0 && ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil {
			if code == "" {
				code = string(CodeOf(Wrap(child, err)))
			}
			return nil
		}
		if entry.IsDir() {
			if child != path {
				event.Dirs++
			}
			return nil
		}
		info, statErr := entry.Info()
		if statErr != nil {
			if code == "" {
				code = string(CodeOf(Wrap(child, statErr)))
			}
			return nil
		}
		event.Files++
		event.Bytes += info.Size()
		if time.Now().After(*next) {
			*next = time.Now().Add(sizeInterval)
			sc.svc.publish(sizeTopic, *event)
		}
		return nil
	})
	if walkErr != nil && code == "" && ctx.Err() == nil {
		code = string(CodeOf(Wrap(path, walkErr)))
	}
	return code
}

func (s *Service) unregisterScan(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.scans, id)
}
