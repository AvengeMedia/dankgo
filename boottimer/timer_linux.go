package boottimer

import (
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func New(d time.Duration) *Timer {
	t, err := newBoottime(d)
	if err != nil {
		return newStd(d)
	}
	return t
}

func newBoottime(d time.Duration) (*Timer, error) {
	fd, err := unix.TimerfdCreate(unix.CLOCK_BOOTTIME, unix.TFD_CLOEXEC|unix.TFD_NONBLOCK)
	if err != nil {
		return nil, err
	}

	// TFD_NONBLOCK makes os.NewFile register the fd with the runtime poller
	// (see os.NewFile), so the reader goroutine parks in Read.
	file := os.NewFile(uintptr(fd), "boottimer")
	c := make(chan time.Time, 1)
	go func() {
		buf := make([]byte, 8)
		for {
			if _, err := file.Read(buf); err != nil {
				return
			}
			select {
			case c <- time.Now():
			default:
			}
		}
	}()

	t := &Timer{C: c}
	t.reset = func(d time.Duration) {
		select {
		case <-c:
		default:
		}
		arm(fd, d)
	}
	t.stop = func() { file.Close() }
	arm(fd, d)
	return t, nil
}

// arm floors the delay at one nanosecond: per timerfd_settime(2), an all-zero
// itimerspec disarms the timer instead of firing it immediately.
func arm(fd int, d time.Duration) {
	if d <= 0 {
		d = time.Nanosecond
	}
	spec := unix.ItimerSpec{Value: unix.NsecToTimespec(d.Nanoseconds())}
	_ = unix.TimerfdSettime(fd, 0, &spec, nil)
}
