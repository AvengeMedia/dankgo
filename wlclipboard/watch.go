//go:build linux

package wlclipboard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/AvengeMedia/dankgo/wayland/ext_data_control"
)

type Change struct {
	Data     []byte
	MimeType string
}

// Watch subscribes to clipboard selection changes and invokes callback with
// the content of each new selection until ctx is done.
func Watch(ctx context.Context, callback func(data []byte, mimeType string)) error {
	s, err := connectSession()
	if err != nil {
		return err
	}
	defer s.Close()

	dataControlMgr, err := s.requireDataControl()
	if err != nil {
		return err
	}

	device, err := dataControlMgr.GetDataDevice(s.seat)
	if err != nil {
		return fmt.Errorf("get data device: %w", err)
	}
	defer device.Destroy()

	offerMimeTypes := make(map[*ext_data_control.ExtDataControlOfferV1][]string)

	device.SetDataOfferHandler(func(e ext_data_control.ExtDataControlDeviceV1DataOfferEvent) {
		if e.Id == nil {
			return
		}
		offerMimeTypes[e.Id] = nil
		e.Id.SetOfferHandler(func(me ext_data_control.ExtDataControlOfferV1OfferEvent) {
			offerMimeTypes[e.Id] = append(offerMimeTypes[e.Id], me.MimeType)
		})
	})

	device.SetSelectionHandler(func(e ext_data_control.ExtDataControlDeviceV1SelectionEvent) {
		if e.Id == nil {
			return
		}

		selectedMime := selectPreferredMimeType(offerMimeTypes[e.Id])
		if selectedMime == "" {
			return
		}

		r, w, err := os.Pipe()
		if err != nil {
			return
		}

		if err := e.Id.Receive(selectedMime, int(w.Fd())); err != nil {
			w.Close()
			r.Close()
			return
		}
		w.Close()

		go func() {
			defer r.Close()
			data, err := io.ReadAll(r)
			if err != nil || len(data) == 0 {
				return
			}
			callback(data, selectedMime)
		}()
	})

	s.display.Roundtrip()
	s.display.Roundtrip()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.ctx.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
				return fmt.Errorf("set read deadline: %w", err)
			}
			if err := s.ctx.Dispatch(); err != nil {
				if isTimeoutError(err) {
					continue
				}
				return fmt.Errorf("dispatch: %w", err)
			}
		}
	}
}

// WatchChan runs Watch on a goroutine, delivering each change on the returned
// channel; a watch failure other than cancellation arrives on the error channel.
func WatchChan(ctx context.Context) (<-chan Change, <-chan error) {
	ch := make(chan Change, 16)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		err := Watch(ctx, func(data []byte, mimeType string) {
			select {
			case ch <- Change{Data: data, MimeType: mimeType}:
			default:
			}
		})
		if err != nil && err != context.Canceled {
			errCh <- err
		}
		close(errCh)
	}()

	time.Sleep(50 * time.Millisecond)
	return ch, errCh
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}
	if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
		return true
	}
	return false
}
