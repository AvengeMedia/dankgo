package httpapi

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/AvengeMedia/dankgo/log"
)

type Server struct {
	server   *http.Server
	listener net.Listener
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Listen() error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	log.Infof("http listening on %s", listener.Addr())
	return nil
}

func (s *Server) Addr() string {
	if s.listener == nil {
		return s.server.Addr
	}
	return s.listener.Addr().String()
}

func (s *Server) Serve(ctx context.Context) error {
	if s.listener == nil {
		if err := s.Listen(); err != nil {
			return err
		}
	}

	errCh := make(chan error, 1)
	go func() {
		err := s.server.Serve(s.listener)
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	}
}

func (s *Server) Run(ctx context.Context) error { return s.Serve(ctx) }

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
