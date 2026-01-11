package app

import (
	"context"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type App struct {
	logger  Logger
	storage storage.Storage
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

func New(logger Logger, storage storage.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	a.logger.Info("create event with ID " + event.ID)
	return a.storage.Create(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, eventID string, event storage.Event) error {
	a.logger.Info("Update event with ID " + eventID)
	return a.storage.Update(ctx, eventID, event)
}

func (a *App) DeleteEvent(ctx context.Context, userID uint64, eventID string) error {
	a.logger.Info("Delete event with ID " + eventID)
	return a.storage.Delete(ctx, userID, eventID)
}

func (a *App) ListDay(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	a.logger.Info("List event on " + date.String())
	return a.storage.ListDay(ctx, userID, date)
}

func (a *App) ListWeek(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	a.logger.Info("List event on " + date.String())
	return a.storage.ListWeek(ctx, userID, date)
}

func (a *App) ListMonth(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	a.logger.Info("List event on " + date.String())
	return a.storage.ListMonth(ctx, userID, date)
}
