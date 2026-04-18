package app

import (
	"context"
	"time"

	"github.com/Evgeny2015/OTUS-Go-HomeWork/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Warning(msg string)
	Debug(msg string)
}

type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (*storage.Event, error)
	ListEvents(ctx context.Context) ([]storage.Event, error)
	ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error)
	ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.UpdateEvent(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
	return a.storage.GetEvent(ctx, id)
}

func (a *App) ListEvents(ctx context.Context) ([]storage.Event, error) {
	return a.storage.ListEvents(ctx)
}

func (a *App) ListEventsForDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForDay(ctx, date)
}

func (a *App) ListEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForWeek(ctx, startOfWeek)
}

func (a *App) ListEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]storage.Event, error) {
	return a.storage.ListEventsForMonth(ctx, startOfMonth)
}
