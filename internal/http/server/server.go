package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

type HTTPServer struct {
	server *http.Server
	logs   *zap.SugaredLogger
}

func NewHTTP(logger *zap.SugaredLogger, mux http.Handler, port string) *HTTPServer {
	server := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%s", port),
	}
	return &HTTPServer{
		server: server,
		logs:   logger,
	}
}

func (s *HTTPServer) Run(done chan<- os.Signal) {
	s.logs.Info("server starting...")
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logs.Errorw("server shut down unexpectedly", "error", err)
	}
	done <- nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logs.Info("shutting down server...")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logs.Error(
			"server shutdown failed",
			"error", err,
		)
		return err
	}
	return nil
}
