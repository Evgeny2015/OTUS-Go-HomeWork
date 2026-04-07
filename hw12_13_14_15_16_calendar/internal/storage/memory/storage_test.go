package memorystorage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	ctx := context.Background()
	s := New()

	// Create event
	event := storage.Event{
		ID:           "test-id",
		Title:        "Test Event",
		DateTime:     time.Now(),
		Duration:     time.Hour,
		Description:  "Test description",
		UserID:       "user1",
		NotifyBefore: time.Minute * 30,
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	// Get event
	retrieved, err := s.GetEvent(ctx, "test-id")
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}
	if retrieved.ID != event.ID || retrieved.Title != event.Title {
		t.Errorf("GetEvent returned wrong event: got %+v, want %+v", retrieved, event)
	}

	// Update event
	updatedEvent := event
	updatedEvent.Title = "Updated Title"
	err = s.UpdateEvent(ctx, updatedEvent)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	retrieved, err = s.GetEvent(ctx, "test-id")
	if err != nil {
		t.Fatalf("GetEvent after update failed: %v", err)
	}
	if retrieved.Title != "Updated Title" {
		t.Errorf("UpdateEvent didn't update title: got %s", retrieved.Title)
	}

	// List events
	events, err := s.ListEvents(ctx)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("ListEvents returned %d events, expected 1", len(events))
	}

	// Delete event
	err = s.DeleteEvent(ctx, "test-id")
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	// Ensure event is deleted
	_, err = s.GetEvent(ctx, "test-id")
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound after delete, got %v", err)
	}
}

func TestStorageDuplicateCreate(t *testing.T) {
	ctx := context.Background()
	s := New()
	event := storage.Event{ID: "dup", Title: "Duplicate"}
	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("First CreateEvent failed: %v", err)
	}
	err = s.CreateEvent(ctx, event)
	if !errors.Is(err, storage.ErrEventAlreadyExists) {
		t.Errorf("Expected ErrEventAlreadyExists, got %v", err)
	}
}

func TestStorageUpdateNonExistent(t *testing.T) {
	ctx := context.Background()
	s := New()
	event := storage.Event{ID: "nonexistent", Title: "No"}
	err := s.UpdateEvent(ctx, event)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestStorageDeleteNonExistent(t *testing.T) {
	ctx := context.Background()
	s := New()
	err := s.DeleteEvent(ctx, "nonexistent")
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}
