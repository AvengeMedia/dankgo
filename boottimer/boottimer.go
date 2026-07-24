// Package boottimer provides a one-shot timer whose deadline keeps counting
// across system suspend. On Linux the Go runtime's monotonic clock pauses
// while the system sleeps, so timers fire late by the suspended duration
// (https://github.com/golang/go/issues/24595); the timer is backed by a
// CLOCK_BOOTTIME timerfd there. Other platforms fall back to a standard
// timer: exact on FreeBSD, whose CLOCK_MONOTONIC counts suspended time (see
// clock_gettime(2)), best-effort where the clock pauses, such as macOS
// (https://github.com/golang/go/issues/66870).
package boottimer

import "time"

// Timer fires on C once per arming. Reset re-arms and drains any undelivered
// fire, but a fire racing a concurrent Reset can still slip through, so treat
// C as a wake-up hint and re-check the deadline.
type Timer struct {
	C <-chan time.Time

	reset func(time.Duration)
	stop  func()
}

func (t *Timer) Reset(d time.Duration) { t.reset(d) }

// Stop releases the timer's resources; the timer is unusable afterwards.
func (t *Timer) Stop() { t.stop() }

func newStd(d time.Duration) *Timer {
	inner := time.NewTimer(d)
	return &Timer{
		C:     inner.C,
		reset: func(d time.Duration) { inner.Reset(d) },
		stop:  func() { inner.Stop() },
	}
}
