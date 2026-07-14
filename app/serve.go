package app

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/AvengeMedia/dankgo/log"
	"golang.org/x/sync/errgroup"
)

const shutdownTimeout = 10 * time.Second

type Runner interface {
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type runnerFunc struct {
	run func(ctx context.Context) error
}

func (r runnerFunc) Run(ctx context.Context) error { return r.run(ctx) }

func (r runnerFunc) Shutdown(context.Context) error { return nil }

func RunnerFunc(run func(ctx context.Context) error) Runner {
	return runnerFunc{run: run}
}

func Serve(ctx context.Context, runners ...Runner) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, gctx := errgroup.WithContext(ctx)
	for _, r := range runners {
		g.Go(func() error { return r.Run(gctx) })
	}

	<-gctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	for _, r := range runners {
		if err := r.Shutdown(shutdownCtx); err != nil {
			log.Warnf("shutdown: %v", err)
		}
	}
	return g.Wait()
}
