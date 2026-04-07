package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func New() *Storage {
	return &Storage{}
}

// NewWithDSN creates a new Storage and connects to the database using the provided DSN.
func NewWithDSN(dsn string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	// For backward compatibility, use default DSN
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	s.db = db
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	query := `
		INSERT INTO events (id, title, date_time, duration, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.DateTime,
		event.Duration.Nanoseconds(),
		event.Description,
		event.UserID,
		event.NotifyBefore.Nanoseconds(),
	)
	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return storage.ErrEventAlreadyExists
		}
		return fmt.Errorf("failed to create event: %w", err)
	}
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	query := `
		UPDATE events
		SET title = $2,
			date_time = $3,
			duration = $4,
			description = $5,
			user_id = $6,
			notify_before = $7,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Title,
		event.DateTime,
		event.Duration.Nanoseconds(),
		event.Description,
		event.UserID,
		event.NotifyBefore.Nanoseconds(),
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
	query := `
		SELECT id, title, date_time, duration, description, user_id, notify_before
		FROM events
		WHERE id = $1
	`
	var event dbEvent
	err := s.db.GetContext(ctx, &event, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event.toStorageEvent(), nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	query := `
		SELECT id, title, date_time, duration, description, user_id, notify_before
		FROM events
		ORDER BY date_time
	`
	var dbEvents []dbEvent
	err := s.db.SelectContext(ctx, &dbEvents, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	events := make([]storage.Event, len(dbEvents))
	for i, e := range dbEvents {
		events[i] = *e.toStorageEvent()
	}
	return events, nil
}

func (s *Storage) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	return s.listEventsInRange(ctx, startOfDay, endOfDay)
}

func (s *Storage) ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error) {
	// Normalize to start of week (Monday)
	weekStart := time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.Add(-24 * time.Hour)
	}
	weekEnd := weekStart.Add(7 * 24 * time.Hour)
	return s.listEventsInRange(ctx, weekStart, weekEnd)
}

func (s *Storage) ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error) {
	// First day of the month
	monthStart := time.Date(startOfMonth.Year(), startOfMonth.Month(), 1, 0, 0, 0, 0, startOfMonth.Location())
	// First day of next month
	monthEnd := monthStart.AddDate(0, 1, 0)
	return s.listEventsInRange(ctx, monthStart, monthEnd)
}

func (s *Storage) listEventsInRange(ctx context.Context, start, end time.Time) ([]storage.Event, error) {
	query := `
		SELECT id, title, date_time, duration, description, user_id, notify_before
		FROM events
		WHERE date_time >= $1 AND date_time < $2
		ORDER BY date_time
	`
	var dbEvents []dbEvent
	err := s.db.SelectContext(ctx, &dbEvents, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to list events in range: %w", err)
	}
	events := make([]storage.Event, len(dbEvents))
	for i, e := range dbEvents {
		events[i] = *e.toStorageEvent()
	}
	return events, nil
}

// dbEvent is a helper struct for database mapping
type dbEvent struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	DateTime     time.Time `db:"date_time"`
	Duration     int64     `db:"duration"` // nanoseconds
	Description  string    `db:"description"`
	UserID       string    `db:"user_id"`
	NotifyBefore *int64    `db:"notify_before"` // nullable
}

func (e *dbEvent) toStorageEvent() *storage.Event {
	event := &storage.Event{
		ID:          e.ID,
		Title:       e.Title,
		DateTime:    e.DateTime,
		Duration:    time.Duration(e.Duration),
		Description: e.Description,
		UserID:      e.UserID,
	}
	if e.NotifyBefore != nil {
		event.NotifyBefore = time.Duration(*e.NotifyBefore)
	}
	return event
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL duplicate key error code is 23505
	// Check if error message contains duplicate key violation
	errStr := err.Error()
	return strings.Contains(errStr, "23505") || strings.Contains(errStr, "duplicate key")
}
