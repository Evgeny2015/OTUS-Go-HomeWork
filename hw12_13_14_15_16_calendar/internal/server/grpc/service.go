package grpcserver

import (
	"context"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/server/protobuf"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// calendarServiceServer implements the gRPC CalendarService
type calendarServiceServer struct {
	protobuf.UnimplementedCalendarServiceServer
	app    Application
	logger Logger
}

// NewCalendarServiceServer creates a new gRPC calendar service server
func NewCalendarServiceServer(app Application, logger Logger) protobuf.CalendarServiceServer {
	return &calendarServiceServer{
		app:    app,
		logger: logger,
	}
}

// CreateEvent implements CreateEvent RPC
func (s *calendarServiceServer) CreateEvent(ctx context.Context, req *protobuf.CreateEventRequest) (*protobuf.CreateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	event, err := protoToStorageEvent(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.app.CreateEvent(ctx, event); err != nil {
		s.logger.Error("CreateEvent failed: " + err.Error())
		return nil, status.Error(codes.Internal, "failed to create event")
	}

	return &protobuf.CreateEventResponse{
		Event: storageToProtoEvent(&event),
	}, nil
}

// UpdateEvent implements UpdateEvent RPC
func (s *calendarServiceServer) UpdateEvent(ctx context.Context, req *protobuf.UpdateEventRequest) (*protobuf.UpdateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	event, err := protoToStorageEvent(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Ensure the event ID matches the request ID
	event.ID = req.GetId()

	if err := s.app.UpdateEvent(ctx, event); err != nil {
		s.logger.Error("UpdateEvent failed: " + err.Error())
		if err == storage.ErrEventNotFound {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, "failed to update event")
	}

	return &protobuf.UpdateEventResponse{
		Event: storageToProtoEvent(&event),
	}, nil
}

// DeleteEvent implements DeleteEvent RPC
func (s *calendarServiceServer) DeleteEvent(ctx context.Context, req *protobuf.DeleteEventRequest) (*protobuf.DeleteEventResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.app.DeleteEvent(ctx, req.GetId()); err != nil {
		s.logger.Error("DeleteEvent failed: " + err.Error())
		if err == storage.ErrEventNotFound {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete event")
	}

	return &protobuf.DeleteEventResponse{
		Success: true,
	}, nil
}

// ListEventsForDay implements ListEventsForDay RPC
func (s *calendarServiceServer) ListEventsForDay(ctx context.Context, req *protobuf.ListEventsForDayRequest) (*protobuf.ListEventsForDayResponse, error) {
	date, err := protoToTime(req.GetDate())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	events, err := s.app.ListEventsForDay(ctx, date)
	if err != nil {
		s.logger.Error("ListEventsForDay failed: " + err.Error())
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	return &protobuf.ListEventsForDayResponse{
		Events: storageToProtoEvents(events),
	}, nil
}

// ListEventsForWeek implements ListEventsForWeek RPC
func (s *calendarServiceServer) ListEventsForWeek(ctx context.Context, req *protobuf.ListEventsForWeekRequest) (*protobuf.ListEventsForWeekResponse, error) {
	startOfWeek, err := protoToTime(req.GetStartOfWeek())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start of week date")
	}

	events, err := s.app.ListEventsForWeek(ctx, startOfWeek)
	if err != nil {
		s.logger.Error("ListEventsForWeek failed: " + err.Error())
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	return &protobuf.ListEventsForWeekResponse{
		Events: storageToProtoEvents(events),
	}, nil
}

// ListEventsForMonth implements ListEventsForMonth RPC
func (s *calendarServiceServer) ListEventsForMonth(ctx context.Context, req *protobuf.ListEventsForMonthRequest) (*protobuf.ListEventsForMonthResponse, error) {
	startOfMonth, err := protoToTime(req.GetStartOfMonth())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start of month date")
	}

	events, err := s.app.ListEventsForMonth(ctx, startOfMonth)
	if err != nil {
		s.logger.Error("ListEventsForMonth failed: " + err.Error())
		return nil, status.Error(codes.Internal, "failed to list events")
	}

	return &protobuf.ListEventsForMonthResponse{
		Events: storageToProtoEvents(events),
	}, nil
}
