package boottimer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFires(t *testing.T) {
	timer := New(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-time.After(time.Second):
		t.Fatal("timer did not fire")
	}
}

func TestResetImmediate(t *testing.T) {
	timer := New(time.Hour)
	defer timer.Stop()

	timer.Reset(0)
	select {
	case <-timer.C:
	case <-time.After(time.Second):
		t.Fatal("Reset(0) did not fire")
	}
}

func TestResetExtends(t *testing.T) {
	timer := New(10 * time.Millisecond)
	defer timer.Stop()

	timer.Reset(time.Hour)
	select {
	case <-timer.C:
		t.Fatal("fired despite extended deadline")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestResetDrainsPendingFire(t *testing.T) {
	timer := New(time.Millisecond)
	defer timer.Stop()

	time.Sleep(50 * time.Millisecond)
	timer.Reset(30 * time.Millisecond)

	start := time.Now()
	select {
	case <-timer.C:
		assert.GreaterOrEqual(t, time.Since(start), 20*time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("re-armed timer did not fire")
	}
}

func TestStdFallbackMatchesBehavior(t *testing.T) {
	timer := newStd(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-time.After(time.Second):
		t.Fatal("std timer did not fire")
	}

	timer.Reset(time.Hour)
	select {
	case <-timer.C:
		t.Fatal("std timer fired despite extended deadline")
	case <-time.After(50 * time.Millisecond):
	}

	timer.Reset(0)
	require.Eventually(t, func() bool {
		select {
		case <-timer.C:
			return true
		default:
			return false
		}
	}, time.Second, 5*time.Millisecond)
}
