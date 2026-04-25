package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
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
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (*storage.Event, error)
	ListEvents(ctx context.Context) ([]storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, host, port string) *Server {
	mux := http.NewServeMux()

	handlers := NewHandlers(app)

	// Register API endpoints
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateEventHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/events/", func(w http.ResponseWriter, r *http.Request) {
		// For now, handle update and delete via request body
		// In a more RESTful design, we'd extract ID from URL path
		switch r.Method {
		case http.MethodPut:
			handlers.UpdateEventHandler(w, r)
		case http.MethodDelete:
			handlers.DeleteEventHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/events/day", handlers.ListEventsForDayHandler)
	mux.HandleFunc("/events/week", handlers.ListEventsForWeekHandler)
	mux.HandleFunc("/events/month", handlers.ListEventsForMonthHandler)

	// Keep hello-world endpoint for root path
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
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello, World!")
	})

	// Add /health endpoint for health checks
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Wrap with logging middleware
	handler := loggingMiddleware(logger, mux)

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
