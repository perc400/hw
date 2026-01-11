package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type Server struct {
	server *http.Server
	logger Logger
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, eventID string, event storage.Event) error
	DeleteEvent(ctx context.Context, userID uint64, eventID string) error
	ListDay(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error)
	ListWeek(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error)
	ListMonth(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, host string, port string) *Server {
	mux := http.NewServeMux()

	handlers := NewHandler(app, logger)
	mux.HandleFunc("/events", handlers.events)
	mux.HandleFunc("/events/day", handlers.listDayEvents)
	mux.HandleFunc("/events/week", handlers.listWeekEvents)
	mux.HandleFunc("/events/month", handlers.listMonthEvents)

	handler := loggingMiddleware(logger, mux)

	return &Server{
		logger: logger,
		server: &http.Server{
			Addr:              net.JoinHostPort(host, port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second, // G112: Potential Slowloris Attack
		},
	}
}

func NewHandlerForTest(logger Logger, app Application) http.Handler {
	mux := http.NewServeMux()

	handlers := NewHandler(app, logger)
	mux.HandleFunc("/events", handlers.events)
	mux.HandleFunc("/events/day", handlers.listDayEvents)
	mux.HandleFunc("/events/week", handlers.listWeekEvents)
	mux.HandleFunc("/events/month", handlers.listMonthEvents)

	return loggingMiddleware(logger, mux)
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server")

	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server")
	return s.server.Shutdown(ctx)
}
