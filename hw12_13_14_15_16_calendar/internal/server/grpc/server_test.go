package grpcserver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/server/protobuf"
	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mockLogger implements Logger interface for testing.
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Info(msg string) {
	m.messages = append(m.messages, "INFO: "+msg)
}

func (m *mockLogger) Error(msg string) {
	m.messages = append(m.messages, "ERROR: "+msg)
}

func (m *mockLogger) Warning(msg string) {
	m.messages = append(m.messages, "WARNING: "+msg)
}

func (m *mockLogger) Debug(msg string) {
	m.messages = append(m.messages, "DEBUG: "+msg)
}

// mockApp implements Application interface for testing.
type mockApp struct {
	events map[string]storage.Event // in-memory storage for testing
	// error control
	createError      error
	updateError      error
	deleteError      error
	getError         error
	listError        error
	listDayError     error
	listWeekError    error
	listMonthError   error
	lastCalledMethod string
}

func newMockApp() *mockApp {
	return &mockApp{
		events: make(map[string]storage.Event),
	}
}

func (m *mockApp) CreateEvent(ctx context.Context, event storage.Event) error {
	m.lastCalledMethod = "CreateEvent"
	if m.createError != nil {
		return m.createError
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockApp) UpdateEvent(ctx context.Context, event storage.Event) error {
	m.lastCalledMethod = "UpdateEvent"
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.events[event.ID]; !exists {
		return storage.ErrEventNotFound
	}
	m.events[event.ID] = event
	return nil
}

func (m *mockApp) DeleteEvent(ctx context.Context, id string) error {
	m.lastCalledMethod = "DeleteEvent"
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, exists := m.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *mockApp) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
	m.lastCalledMethod = "GetEvent"
	if m.getError != nil {
		return nil, m.getError
	}
	event, exists := m.events[id]
	if !exists {
		return nil, storage.ErrEventNotFound
	}
	return &event, nil
}

func (m *mockApp) ListEvents(ctx context.Context) ([]storage.Event, error) {
	m.lastCalledMethod = "ListEvents"
	if m.listError != nil {
		return nil, m.listError
	}
	events := make([]storage.Event, 0, len(m.events))
	for _, event := range m.events {
		events = append(events, event)
	}
	return events, nil
}

func (m *mockApp) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	m.lastCalledMethod = "ListEventsForDay"
	if m.listDayError != nil {
		return nil, m.listDayError
	}
	events := make([]storage.Event, 0, len(m.events))
	for _, event := range m.events {
		if event.DateTime.Year() == date.Year() && event.DateTime.Month() == date.Month() && event.DateTime.Day() == date.Day() {
			events = append(events, event)
		}
	}
	return events, nil
}

func (m *mockApp) ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error) {
	m.lastCalledMethod = "ListEventsForWeek"
	if m.listWeekError != nil {
		return nil, m.listWeekError
	}
	endOfWeek := startOfWeek.AddDate(0, 0, 7)
	events := make([]storage.Event, 0, len(m.events))
	for _, event := range m.events {
		if !event.DateTime.Before(startOfWeek) && event.DateTime.Before(endOfWeek) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (m *mockApp) ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error) {
	m.lastCalledMethod = "ListEventsForMonth"
	if m.listMonthError != nil {
		return nil, m.listMonthError
	}
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	events := make([]storage.Event, 0, len(m.events))
	for _, event := range m.events {
		if !event.DateTime.Before(startOfMonth) && event.DateTime.Before(endOfMonth) {
			events = append(events, event)
		}
	}
	return events, nil
}

// Helper function to create a test event
func createTestEvent(id string) *protobuf.Event {
	return &protobuf.Event{
		Id:           id,
		Title:        "Test Event",
		DateTime:     timestamppb.New(time.Now().Add(24 * time.Hour)),
		Duration:     durationpb.New(2 * time.Hour),
		Description:  "Test Description",
		UserId:       "user123",
		NotifyBefore: durationpb.New(30 * time.Minute),
	}
}

// Helper function to create a test storage event
func createTestStorageEvent(id string) storage.Event {
	return storage.Event{
		ID:           id,
		Title:        "Test Event",
		DateTime:     time.Now().Add(24 * time.Hour),
		Duration:     2 * time.Hour,
		Description:  "Test Description",
		UserID:       "user123",
		NotifyBefore: 30 * time.Minute,
	}
}

func TestNewServer(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()

	server := NewServer(logger, app)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.server == nil {
		t.Error("server.server is nil")
	}
	if server.logger != logger {
		t.Error("server.logger doesn't match input logger")
	}
	if server.app != app {
		t.Error("server.app doesn't match input app")
	}
}

func TestServer_StartStop(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()

	server := NewServer(logger, app)

	// Test starting server with invalid port (should fail quickly)
	err := server.Start("localhost", "99999") // invalid port
	if err == nil {
		t.Error("Expected error for invalid port, got nil")
	}

	// Test stop with context
	ctx := context.Background()
	err = server.Stop(ctx)
	if err != nil {
		t.Errorf("Stop returned error: %v", err)
	}
}

func TestLoggingInterceptor(t *testing.T) {
	logger := &mockLogger{}
	interceptor := loggingInterceptor(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/calendar.CalendarService/CreateEvent",
	}

	resp, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("interceptor returned error: %v", err)
	}
	if resp != "response" {
		t.Errorf("expected response 'response', got %v", resp)
	}

	// Check that logger was called
	if len(logger.messages) == 0 {
		t.Error("logger was not called")
	} else if !strings.Contains(logger.messages[0], "gRPC request: /calendar.CalendarService/CreateEvent") {
		t.Errorf("logger message doesn't contain expected text: %s", logger.messages[0])
	}
}

func TestProtoToStorageEvent(t *testing.T) {
	protoEvent := createTestEvent("test-id")

	event, err := protoToStorageEvent(protoEvent)
	if err != nil {
		t.Errorf("protoToStorageEvent returned error: %v", err)
	}
	if event.ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", event.ID)
	}
	if event.Title != "Test Event" {
		t.Errorf("expected Title 'Test Event', got %s", event.Title)
	}
	if event.Description != "Test Description" {
		t.Errorf("expected Description 'Test Description', got %s", event.Description)
	}
	if event.UserID != "user123" {
		t.Errorf("expected UserID 'user123', got %s", event.UserID)
	}
	if event.Duration != 2*time.Hour {
		t.Errorf("expected Duration 2h, got %v", event.Duration)
	}
	if event.NotifyBefore != 30*time.Minute {
		t.Errorf("expected NotifyBefore 30m, got %v", event.NotifyBefore)
	}

	// Test nil event
	_, err = protoToStorageEvent(nil)
	if err == nil {
		t.Error("expected error for nil event, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}
}

func TestStorageToProtoEvent(t *testing.T) {
	storageEvent := createTestStorageEvent("test-id")

	protoEvent := storageToProtoEvent(&storageEvent)
	if protoEvent == nil {
		t.Fatal("storageToProtoEvent returned nil")
	}
	if protoEvent.Id != "test-id" {
		t.Errorf("expected Id 'test-id', got %s", protoEvent.Id)
	}
	if protoEvent.Title != "Test Event" {
		t.Errorf("expected Title 'Test Event', got %s", protoEvent.Title)
	}
	if protoEvent.Description != "Test Description" {
		t.Errorf("expected Description 'Test Description', got %s", protoEvent.Description)
	}
	if protoEvent.UserId != "user123" {
		t.Errorf("expected UserId 'user123', got %s", protoEvent.UserId)
	}

	// Test nil event
	if storageToProtoEvent(nil) != nil {
		t.Error("expected nil for nil input")
	}
}

func TestStorageToProtoEvents(t *testing.T) {
	storageEvents := []storage.Event{
		createTestStorageEvent("id1"),
		createTestStorageEvent("id2"),
	}

	protoEvents := storageToProtoEvents(storageEvents)
	if len(protoEvents) != 2 {
		t.Fatalf("expected 2 events, got %d", len(protoEvents))
	}
	if protoEvents[0].Id != "id1" {
		t.Errorf("expected first event Id 'id1', got %s", protoEvents[0].Id)
	}
	if protoEvents[1].Id != "id2" {
		t.Errorf("expected second event Id 'id2', got %s", protoEvents[1].Id)
	}
}

func TestProtoToTime(t *testing.T) {
	now := time.Now()
	ts := timestamppb.New(now)

	result, err := protoToTime(ts)
	if err != nil {
		t.Errorf("protoToTime returned error: %v", err)
	}
	if !result.Equal(now) {
		t.Errorf("expected time %v, got %v", now, result)
	}

	// Test nil timestamp
	_, err = protoToTime(nil)
	if err == nil {
		t.Error("expected error for nil timestamp, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}
}

// Service method tests
func TestCalendarService_CreateEvent(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// Test successful creation
	req := &protobuf.CreateEventRequest{
		Event: createTestEvent("test-id"),
	}

	resp, err := service.CreateEvent(context.Background(), req)
	if err != nil {
		t.Errorf("CreateEvent returned error: %v", err)
	}
	if resp.Event == nil {
		t.Fatal("response event is nil")
	}
	if resp.Event.Id != "test-id" {
		t.Errorf("expected event ID 'test-id', got %s", resp.Event.Id)
	}
	if app.lastCalledMethod != "CreateEvent" {
		t.Errorf("expected CreateEvent to be called, got %s", app.lastCalledMethod)
	}
	if len(app.events) != 1 {
		t.Errorf("expected 1 event in storage, got %d", len(app.events))
	}

	// Test nil event
	_, err = service.CreateEvent(context.Background(), &protobuf.CreateEventRequest{})
	if err == nil {
		t.Error("expected error for nil event, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test app error
	app.createError = errors.New("storage error")
	_, err = service.CreateEvent(context.Background(), req)
	if err == nil {
		t.Error("expected error when app.CreateEvent fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}

func TestCalendarService_UpdateEvent(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// First create an event
	storageEvent := createTestStorageEvent("test-id")
	app.events["test-id"] = storageEvent

	// Test successful update
	req := &protobuf.UpdateEventRequest{
		Id: "test-id",
		Event: &protobuf.Event{
			Id:           "test-id",
			Title:        "Updated Title",
			DateTime:     timestamppb.New(time.Now().Add(48 * time.Hour)),
			Duration:     durationpb.New(3 * time.Hour),
			Description:  "Updated Description",
			UserId:       "user456",
			NotifyBefore: durationpb.New(1 * time.Hour),
		},
	}

	resp, err := service.UpdateEvent(context.Background(), req)
	if err != nil {
		t.Errorf("UpdateEvent returned error: %v", err)
	}
	if resp.Event == nil {
		t.Fatal("response event is nil")
	}
	if resp.Event.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", resp.Event.Title)
	}
	if app.lastCalledMethod != "UpdateEvent" {
		t.Errorf("expected UpdateEvent to be called, got %s", app.lastCalledMethod)
	}
	if app.events["test-id"].Title != "Updated Title" {
		t.Errorf("event not updated in storage")
	}

	// Test nil event
	_, err = service.UpdateEvent(context.Background(), &protobuf.UpdateEventRequest{Id: "test-id"})
	if err == nil {
		t.Error("expected error for nil event, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test event not found
	req2 := &protobuf.UpdateEventRequest{
		Id:    "non-existent",
		Event: createTestEvent("non-existent"),
	}
	_, err = service.UpdateEvent(context.Background(), req2)
	if err == nil {
		t.Error("expected error for non-existent event, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound code, got %v", status.Code(err))
	}

	// Test app error
	app.updateError = errors.New("storage error")
	_, err = service.UpdateEvent(context.Background(), req)
	if err == nil {
		t.Error("expected error when app.UpdateEvent fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}

func TestCalendarService_DeleteEvent(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// First create an event
	storageEvent := createTestStorageEvent("test-id")
	app.events["test-id"] = storageEvent

	// Test successful deletion
	req := &protobuf.DeleteEventRequest{
		Id: "test-id",
	}

	resp, err := service.DeleteEvent(context.Background(), req)
	if err != nil {
		t.Errorf("DeleteEvent returned error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success true")
	}
	if app.lastCalledMethod != "DeleteEvent" {
		t.Errorf("expected DeleteEvent to be called, got %s", app.lastCalledMethod)
	}
	if len(app.events) != 0 {
		t.Errorf("expected 0 events in storage after deletion, got %d", len(app.events))
	}

	// Test empty ID
	_, err = service.DeleteEvent(context.Background(), &protobuf.DeleteEventRequest{})
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test event not found
	_, err = service.DeleteEvent(context.Background(), &protobuf.DeleteEventRequest{Id: "non-existent"})
	if err == nil {
		t.Error("expected error for non-existent event, got nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound code, got %v", status.Code(err))
	}

	// Test app error
	app.events["test-id-2"] = createTestStorageEvent("test-id-2")
	app.deleteError = errors.New("storage error")
	_, err = service.DeleteEvent(context.Background(), &protobuf.DeleteEventRequest{Id: "test-id-2"})
	if err == nil {
		t.Error("expected error when app.DeleteEvent fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}

func TestCalendarService_ListEventsForDay(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// Create test events with different dates
	now := time.Now()
	todayEvent := storage.Event{
		ID:           "today",
		Title:        "Today Event",
		DateTime:     now,
		Duration:     time.Hour,
		Description:  "Event today",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	tomorrowEvent := storage.Event{
		ID:           "tomorrow",
		Title:        "Tomorrow Event",
		DateTime:     now.AddDate(0, 0, 1),
		Duration:     time.Hour,
		Description:  "Event tomorrow",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	app.events["today"] = todayEvent
	app.events["tomorrow"] = tomorrowEvent

	// Test successful listing for today
	req := &protobuf.ListEventsForDayRequest{
		Date: timestamppb.New(now),
	}

	resp, err := service.ListEventsForDay(context.Background(), req)
	if err != nil {
		t.Errorf("ListEventsForDay returned error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event for today, got %d", len(resp.Events))
	}
	if resp.Events[0].Id != "today" {
		t.Errorf("expected event ID 'today', got %s", resp.Events[0].Id)
	}
	if app.lastCalledMethod != "ListEventsForDay" {
		t.Errorf("expected ListEventsForDay to be called, got %s", app.lastCalledMethod)
	}

	// Test invalid date
	_, err = service.ListEventsForDay(context.Background(), &protobuf.ListEventsForDayRequest{})
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test app error
	app.listDayError = errors.New("storage error")
	_, err = service.ListEventsForDay(context.Background(), req)
	if err == nil {
		t.Error("expected error when app.ListEventsForDay fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}

func TestCalendarService_ListEventsForWeek(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// Create test events
	now := time.Now()
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// Adjust to Monday if needed (simplified)
	inWeekEvent := storage.Event{
		ID:           "in-week",
		Title:        "In Week Event",
		DateTime:     startOfWeek.AddDate(0, 0, 3), // Wednesday
		Duration:     time.Hour,
		Description:  "Event in week",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	nextWeekEvent := storage.Event{
		ID:           "next-week",
		Title:        "Next Week Event",
		DateTime:     startOfWeek.AddDate(0, 0, 8), // Next week
		Duration:     time.Hour,
		Description:  "Event next week",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	app.events["in-week"] = inWeekEvent
	app.events["next-week"] = nextWeekEvent

	// Test successful listing for week
	req := &protobuf.ListEventsForWeekRequest{
		StartOfWeek: timestamppb.New(startOfWeek),
	}

	resp, err := service.ListEventsForWeek(context.Background(), req)
	if err != nil {
		t.Errorf("ListEventsForWeek returned error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event for week, got %d", len(resp.Events))
	}
	if resp.Events[0].Id != "in-week" {
		t.Errorf("expected event ID 'in-week', got %s", resp.Events[0].Id)
	}
	if app.lastCalledMethod != "ListEventsForWeek" {
		t.Errorf("expected ListEventsForWeek to be called, got %s", app.lastCalledMethod)
	}

	// Test invalid date
	_, err = service.ListEventsForWeek(context.Background(), &protobuf.ListEventsForWeekRequest{})
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test app error
	app.listWeekError = errors.New("storage error")
	_, err = service.ListEventsForWeek(context.Background(), req)
	if err == nil {
		t.Error("expected error when app.ListEventsForWeek fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}

func TestCalendarService_ListEventsForMonth(t *testing.T) {
	logger := &mockLogger{}
	app := newMockApp()
	service := NewCalendarServiceServer(app, logger)

	// Create test events with specific dates
	// Use a fixed date to avoid month boundary issues
	startOfMonth := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // January 1, 2024

	inMonthEvent := storage.Event{
		ID:           "in-month",
		Title:        "In Month Event",
		DateTime:     startOfMonth.AddDate(0, 0, 15), // January 16, 2024
		Duration:     time.Hour,
		Description:  "Event in month",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	nextMonthEvent := storage.Event{
		ID:           "next-month",
		Title:        "Next Month Event",
		DateTime:     startOfMonth.AddDate(0, 1, 5), // February 6, 2024 (definitely next month)
		Duration:     time.Hour,
		Description:  "Event next month",
		UserID:       "user1",
		NotifyBefore: 0,
	}
	app.events["in-month"] = inMonthEvent
	app.events["next-month"] = nextMonthEvent

	// Test successful listing for month
	req := &protobuf.ListEventsForMonthRequest{
		StartOfMonth: timestamppb.New(startOfMonth),
	}

	resp, err := service.ListEventsForMonth(context.Background(), req)
	if err != nil {
		t.Errorf("ListEventsForMonth returned error: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Errorf("expected 1 event for month, got %d", len(resp.Events))
	}
	if resp.Events[0].Id != "in-month" {
		t.Errorf("expected event ID 'in-month', got %s", resp.Events[0].Id)
	}
	if app.lastCalledMethod != "ListEventsForMonth" {
		t.Errorf("expected ListEventsForMonth to be called, got %s", app.lastCalledMethod)
	}

	// Test invalid date
	_, err = service.ListEventsForMonth(context.Background(), &protobuf.ListEventsForMonthRequest{})
	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument code, got %v", status.Code(err))
	}

	// Test app error
	app.listMonthError = errors.New("storage error")
	_, err = service.ListEventsForMonth(context.Background(), req)
	if err == nil {
		t.Error("expected error when app.ListEventsForMonth fails, got nil")
	}
	if status.Code(err) != codes.Internal {
		t.Errorf("expected Internal code, got %v", status.Code(err))
	}
}
