//go:build linux

package wlclipboard

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/AvengeMedia/dankgo/wayland/ext_data_control"
)

// ErrOwnerClosed reports a Set against an Owner whose connection has been
// closed or lost.
var ErrOwnerClosed = errors.New("clipboard owner closed")

// Owner holds one long-lived Wayland connection and serves clipboard offers
// from a background goroutine. Each Set replaces the selection; when another
// client takes the selection the Owner stops serving that offer but stays
// usable for the next Set.
type Owner struct {
	session *session
	mgr     *ext_data_control.ExtDataControlManagerV1
	device  *ext_data_control.ExtDataControlDeviceV1
	current *ext_data_control.ExtDataControlSourceV1

	sets     chan setRequest
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

type setRequest struct {
	offers []Offer
	reply  chan error
}

func NewOwner() (*Owner, error) {
	s, err := connectSession()
	if err != nil {
		return nil, err
	}

	mgr, err := s.requireDataControl()
	if err != nil {
		s.Close()
		return nil, err
	}

	device, err := mgr.GetDataDevice(s.seat)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("get data device: %w", err)
	}

	o := &Owner{
		session: s,
		mgr:     mgr,
		device:  device,
		sets:    make(chan setRequest),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	go o.run()
	return o, nil
}

// Set takes the clipboard selection and serves the given offers until another
// client takes it, Set is called again, or the Owner is closed.
func (o *Owner) Set(offers []Offer) error {
	if len(offers) == 0 {
		return errors.New("no offers to serve")
	}

	req := setRequest{offers: offers, reply: make(chan error, 1)}
	select {
	case o.sets <- req:
		return <-req.reply
	case <-o.done:
		return ErrOwnerClosed
	}
}

// Close releases the selection if still held and shuts down the connection.
func (o *Owner) Close() {
	o.stopOnce.Do(func() { close(o.stop) })
	<-o.done
}

// run is the only goroutine touching the Wayland connection; Set requests are
// funneled here so protocol writes never race the dispatch loop.
func (o *Owner) run() {
	defer close(o.done)
	defer o.session.Close()
	defer o.device.Destroy()

	for {
		select {
		case <-o.stop:
			o.dropCurrent()
			return
		case req := <-o.sets:
			req.reply <- o.serve(req.offers)
		default:
			if err := o.session.ctx.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
				return
			}
			if err := o.session.ctx.Dispatch(); err != nil {
				if isTimeoutError(err) {
					continue
				}
				return
			}
		}
	}
}

func (o *Owner) serve(offers []Offer) error {
	source, err := o.mgr.CreateDataSource()
	if err != nil {
		return fmt.Errorf("create data source: %w", err)
	}

	offerData := make(map[string][]byte, len(offers))
	for _, offer := range offers {
		if err := source.Offer(offer.MimeType); err != nil {
			source.Destroy()
			return fmt.Errorf("offer %s: %w", offer.MimeType, err)
		}
		offerData[offer.MimeType] = offer.Data
	}

	source.SetSendHandler(func(e ext_data_control.ExtDataControlSourceV1SendEvent) {
		_ = syscall.SetNonblock(e.Fd, false)
		file := os.NewFile(uintptr(e.Fd), "pipe")
		defer file.Close()

		if data, ok := offerData[e.MimeType]; ok {
			_, _ = file.Write(data)
		}
	})

	cancelled := false
	source.SetCancelledHandler(func(ext_data_control.ExtDataControlSourceV1CancelledEvent) {
		cancelled = true
		source.Destroy()
		if o.current == source {
			o.current = nil
		}
	})

	if err := o.device.SetSelection(source); err != nil {
		source.Destroy()
		return fmt.Errorf("set selection: %w", err)
	}

	if err := o.session.ctx.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("clear read deadline: %w", err)
	}
	o.session.display.Roundtrip()

	if !cancelled {
		o.current = source
	}
	return nil
}

func (o *Owner) dropCurrent() {
	if o.current == nil {
		return
	}
	o.current.Destroy()
	o.current = nil
}
