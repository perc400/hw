package storage

import (
	"context"
	"time"
)

type Storage interface {
	Create(ctx context.Context, event Event) error
	Update(ctx context.Context, eventID string, event Event) error
	Delete(ctx context.Context, eventID string) error
	ListDay(ctx context.Context, date time.Time) ([]Event, error)
	ListWeek(ctx context.Context, date time.Time) ([]Event, error)
	ListMonth(ctx context.Context, date time.Time) ([]Event, error)
}
