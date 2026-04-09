package grpcserver

import (
	"context"
	"net"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/server/protobuf"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	server *grpc.Server
	logger Logger
	app    Application
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

func NewServer(logger Logger, app Application) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	s := &Server{
		server: grpcServer,
		logger: logger,
		app:    app,
	}

	// Register the service implementation
	calendarService := NewCalendarServiceServer(app, logger)
	protobuf.RegisterCalendarServiceServer(grpcServer, calendarService)

	return s
}

func (s *Server) Start(host, port string) error {
	addr := net.JoinHostPort(host, port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("gRPC server starting on " + addr)

	return s.server.Serve(lis)
}

func (s *Server) Stop(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}

// loggingInterceptor creates a gRPC unary interceptor for logging requests
func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		logger.Info("gRPC request: " + info.FullMethod)
		return handler(ctx, req)
	}
}

// Helper functions to convert between protobuf and storage types
func protoToStorageEvent(protoEvent *protobuf.Event) (storage.Event, error) {
	if protoEvent == nil {
		return storage.Event{}, status.Error(codes.InvalidArgument, "event is nil")
	}

	dateTime, err := protoToTime(protoEvent.GetDateTime())
	if err != nil {
		return storage.Event{}, err
	}

	return storage.Event{
		ID:           protoEvent.GetId(),
		Title:        protoEvent.GetTitle(),
		DateTime:     dateTime,
		Duration:     protoEvent.GetDuration().AsDuration(),
		Description:  protoEvent.GetDescription(),
		UserID:       protoEvent.GetUserId(),
		NotifyBefore: protoEvent.GetNotifyBefore().AsDuration(),
	}, nil
}

func storageToProtoEvent(event *storage.Event) *protobuf.Event {
	if event == nil {
		return nil
	}

	return &protobuf.Event{
		Id:           event.ID,
		Title:        event.Title,
		DateTime:     timestamppb.New(event.DateTime),
		Duration:     durationpb.New(event.Duration),
		Description:  event.Description,
		UserId:       event.UserID,
		NotifyBefore: durationpb.New(event.NotifyBefore),
	}
}

func storageToProtoEvents(events []storage.Event) []*protobuf.Event {
	protoEvents := make([]*protobuf.Event, 0, len(events))
	for _, event := range events {
		protoEvents = append(protoEvents, storageToProtoEvent(&event))
	}
	return protoEvents
}

func protoToTime(ts *timestamppb.Timestamp) (time.Time, error) {
	if ts == nil {
		return time.Time{}, status.Error(codes.InvalidArgument, "timestamp is required")
	}
	return ts.AsTime(), nil
}
