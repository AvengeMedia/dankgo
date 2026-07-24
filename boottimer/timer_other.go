//go:build !linux

package boottimer

import "time"

func New(d time.Duration) *Timer {
	return newStd(d)
}
