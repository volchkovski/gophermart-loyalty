package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/volchkovski/gophermart-loyalty/internal/logger"
)

type HTTPServer struct {
	server *http.Server
	notify chan error
}

func New(addr string, router http.Handler) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:           addr,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    30 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1 MB
		},
		notify: make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		defer close(s.notify)

		logger.Log.Infof("HTTP server starting on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.notify <- err
			return
		}
	}()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	logger.Log.Info("HTTP server shutting down gracefully...")

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Log.Errorf("HTTP server shutdown error: %s", err)
		return err
	}

	logger.Log.Info("HTTP server stopped gracefully")
	return nil
}

func (s *HTTPServer) Notify() chan error {
	return s.notify
}
