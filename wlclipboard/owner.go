//go:build unix

package wlclipboard

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/AvengeMedia/dankgo/wayland/ext_data_control"
)

// ErrOwnerClosed reports a Set against an Owner whose connection has been
// closed or lost.
var ErrOwnerClosed = errors.New("clipboard owner closed")

// Owner holds one long-lived Wayland connection and serves clipboard offers
// from a background goroutine. Each Set replaces the selection; when another
// client takes the selection the Owner stops serving that offer but stays
// usable for the next Set.
//
// The goroutine parks in poll(2) on the Wayland connection and a wakeup pipe,
// so an Owner that nobody is asking anything of costs no CPU at all, and an
// offer request from another client is served as soon as it arrives.
type Owner struct {
	session *session
	mgr     *ext_data_control.ExtDataControlManagerV1
	device  *ext_data_control.ExtDataControlDeviceV1
	current *ext_data_control.ExtDataControlSourceV1

	fd    int
	wakeR int
	wakeW int

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

	wake, err := wakePipe()
	if err != nil {
		device.Destroy()
		s.Close()
		return nil, err
	}

	o := &Owner{
		session: s,
		mgr:     mgr,
		device:  device,
		fd:      s.ctx.Fd(),
		wakeR:   wake[0],
		wakeW:   wake[1],
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
	// Wake first: run is parked in poll and would not reach the channel
	// until the compositor happened to send something.
	o.wake()
	select {
	case o.sets <- req:
		return <-req.reply
	case <-o.done:
		return ErrOwnerClosed
	}
}

// Close releases the selection if still held and shuts down the connection.
func (o *Owner) Close() {
	o.stopOnce.Do(func() {
		close(o.stop)
		o.wake()
	})
	<-o.done
}

// wakePipe is the self-pipe run is woken through. Pipe plus fcntl rather than
// pipe2 because darwin has no pipe2, and nothing here forks between the two.
func wakePipe() ([2]int, error) {
	var fds [2]int
	if err := unix.Pipe(fds[:]); err != nil {
		return fds, fmt.Errorf("wakeup pipe: %w", err)
	}
	for _, fd := range fds {
		unix.CloseOnExec(fd)
		if err := unix.SetNonblock(fd, true); err != nil {
			_ = unix.Close(fds[0])
			_ = unix.Close(fds[1])
			return fds, fmt.Errorf("wakeup pipe: %w", err)
		}
	}
	return fds, nil
}

// wake interrupts the poll in run so a pending Set or Close is picked up
// without waiting for the compositor to send something.
func (o *Owner) wake() {
	var b [1]byte
	for {
		_, err := unix.Write(o.wakeW, b[:])
		if err == unix.EINTR {
			continue
		}
		return
	}
}

// run is the only goroutine touching the Wayland connection; Set requests are
// funneled here so protocol writes never race the dispatch loop.
func (o *Owner) run() {
	defer close(o.done)
	defer o.closeWake()
	defer o.session.Close()
	defer o.device.Destroy()

	for {
		select {
		case <-o.stop:
			o.dropCurrent()
			return
		case req := <-o.sets:
			req.reply <- o.serve(req.offers)
			continue
		default:
		}

		readable, err := o.await()
		if err != nil {
			return
		}
		if !readable {
			continue
		}
		if err := o.session.ctx.Dispatch(); err != nil {
			return
		}
	}
}

// await parks until the compositor sends something or another goroutine calls
// wake, and reports whether the connection has a message to read. Dispatch is
// only safe to call once poll says so: it reads a message header and body in
// two syscalls, so interrupting it partway would desync the stream.
func (o *Owner) await() (bool, error) {
	fds := []unix.PollFd{
		{Fd: int32(o.fd), Events: unix.POLLIN},
		{Fd: int32(o.wakeR), Events: unix.POLLIN},
	}

	for {
		n, err := unix.Poll(fds, -1)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("poll: %w", err)
		}
		if n == 0 {
			continue
		}
		if fds[1].Revents&unix.POLLIN != 0 {
			o.drainWake()
		}
		if fds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return false, errors.New("wayland connection lost")
		}
		return fds[0].Revents&unix.POLLIN != 0, nil
	}
}

func (o *Owner) drainWake() {
	var buf [16]byte
	for {
		n, err := unix.Read(o.wakeR, buf[:])
		if err == unix.EINTR {
			continue
		}
		if err != nil || n < len(buf) {
			return
		}
	}
}

func (o *Owner) closeWake() {
	_ = unix.Close(o.wakeR)
	_ = unix.Close(o.wakeW)
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
