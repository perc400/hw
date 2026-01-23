package internalgrpc

import (
	"context"
	"net"
	"time"

	eventpb "github.com/perc400/hw/hw12_13_14_15_calendar/api/event" //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage"  //nolint:depguard
	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
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

func NewServer(logger Logger, app Application) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor(logger)),
	)

	handler := NewHandler(app, logger)
	eventpb.RegisterEventServiceServer(grpcServer, handler)

	return &Server{
		server: grpcServer,
		logger: logger,
	}
}

func (s *Server) ServeListener(lsn net.Listener) error {
	return s.server.Serve(lsn)
}

func (s *Server) Start(port string) error {
	s.logger.Info("Starting gRPC server")

	lsn, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	return s.server.Serve(lsn)
}

func (s *Server) Stop() {
	s.logger.Info("Shutting down gRPC server")
	s.server.GracefulStop()
}
