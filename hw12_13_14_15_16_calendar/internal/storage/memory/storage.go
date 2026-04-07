package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; exists {
		return storage.ErrEventAlreadyExists
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		return storage.ErrEventNotFound
	}
	s.events[event.ID] = event
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	delete(s.events, id)
	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, storage.ErrEventNotFound
	}
	return &event, nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, event)
	}
	return events, nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events := make([]storage.Event, 0)
	for _, event := range s.events {
		if !event.DateTime.Before(startOfDay) && event.DateTime.Before(endOfDay) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Normalize to start of week (Monday)
	weekStart := time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	// Adjust to Monday if not already
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.Add(-24 * time.Hour)
	}
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	events := make([]storage.Event, 0)
	for _, event := range s.events {
		if !event.DateTime.Before(weekStart) && event.DateTime.Before(weekEnd) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First day of the month
	monthStart := time.Date(startOfMonth.Year(), startOfMonth.Month(), 1, 0, 0, 0, 0, startOfMonth.Location())
	// First day of next month
	monthEnd := monthStart.AddDate(0, 1, 0)

	events := make([]storage.Event, 0)
	for _, event := range s.events {
		if !event.DateTime.Before(monthStart) && event.DateTime.Before(monthEnd) {
			events = append(events, event)
		}
	}
	return events, nil
}
