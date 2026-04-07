package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
	logger Logger
	host   string
	port   string
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Warning(msg string)
	Debug(msg string)
}

type Application interface {
	// This interface is kept for compatibility but not used in this simple server
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	mux := http.NewServeMux()

	// Add hello-world endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})

	// Add /hello endpoint as well
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})

	// Wrap with logging middleware
	handler := loggingMiddleware(mux)

	server := &http.Server{
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		server: server,
		logger: logger,
		host:   host,
		port:   port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.host, s.port)
	s.server.Addr = addr

	s.logger.Info(fmt.Sprintf("HTTP server starting on %s", addr))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(shutdownCtx)
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down")
	return s.server.Shutdown(ctx)
}
