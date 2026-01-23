package storage

import (
	"context"
	"time"
)

type Storage interface {
	Create(ctx context.Context, event Event) error
	Update(ctx context.Context, eventID string, event Event) error
	Delete(ctx context.Context, userID uint64, eventID string) error
	ListDay(ctx context.Context, userID uint64, date time.Time) ([]Event, error)
	ListWeek(ctx context.Context, userID uint64, date time.Time) ([]Event, error)
	ListMonth(ctx context.Context, userID uint64, date time.Time) ([]Event, error)
	MarkNotified(ctx context.Context, eventID string, now time.Time) error
	ListEventsToNotify(ctx context.Context, now time.Time) ([]Event, error)
	DeleteOldEvents(ctx context.Context, before time.Time) error
}
