package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AvengeMedia/dankgo/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeStopsAllOnFirstError(t *testing.T) {
	boom := errors.New("boom")
	otherStopped := make(chan struct{})

	err := app.Serve(context.Background(),
		app.RunnerFunc(func(ctx context.Context) error { return boom }),
		app.RunnerFunc(func(ctx context.Context) error {
			<-ctx.Done()
			close(otherStopped)
			return nil
		}),
	)
	assert.ErrorIs(t, err, boom)

	select {
	case <-otherStopped:
	case <-time.After(time.Second):
		t.Fatal("sibling runner was not cancelled")
	}
}

func TestServeCleanShutdownOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdownCalled := make(chan struct{})

	runner := &trackingRunner{stopped: shutdownCalled}
	done := make(chan error, 1)
	go func() { done <- app.Serve(ctx, runner) }()

	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("Serve did not return after cancel")
	}

	select {
	case <-shutdownCalled:
	case <-time.After(time.Second):
		t.Fatal("Shutdown was not called")
	}
}

type trackingRunner struct {
	stopped chan struct{}
}

func (r *trackingRunner) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (r *trackingRunner) Shutdown(context.Context) error {
	close(r.stopped)
	return nil
}
