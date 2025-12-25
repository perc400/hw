package internalhttp

import (
	"context"
	"net"
	"net/http"
	"time"
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

type Application interface { // TODO
}

//nolint:revive
func NewServer(logger Logger, app Application, host string, port string) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("hello"))
		if err != nil {
			logger.Error("failed to write response" + err.Error())
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

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

// TODO
