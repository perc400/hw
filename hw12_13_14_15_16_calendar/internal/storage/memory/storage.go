package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type Storage struct {
	events map[string]storage.Event
	mu     sync.RWMutex
}

func New() *Storage {
	eventStorage := make(map[string]storage.Event)
	return &Storage{
		events: eventStorage,
	}
}

func hasOverlap(e1, e2 storage.Event) bool {
	end1 := e1.Datetime.Add(e1.Duration)
	end2 := e2.Datetime.Add(e2.Duration)

	return e1.Datetime.Before(end2) && e2.Datetime.Before(end1)
}

func (s *Storage) Create(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = ctx
	if _, exists := s.events[event.ID]; exists {
		return storage.ErrEventAlreadyExists
	}

	for _, existing := range s.events {
		if existing.UserID != event.UserID {
			continue
		}

		if hasOverlap(event, existing) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) Update(ctx context.Context, eventID string, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = ctx
	if _, exists := s.events[eventID]; !exists {
		return storage.ErrEventNotFound
	}

	for _, existing := range s.events {
		if existing.UserID != event.UserID {
			continue
		}

		if existing.ID == s.events[eventID].ID {
			continue
		}

		if hasOverlap(event, existing) {
			return storage.ErrDateBusy
		}
	}

	event.ID = eventID
	s.events[eventID] = event
	return nil
}

func (s *Storage) Delete(ctx context.Context, eventID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = ctx
	if _, exists := s.events[eventID]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, eventID)
	return nil
}

func (s *Storage) ListDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_ = ctx
	dayStart := time.Date(
		date.Year(), date.Month(), date.Day(),
		0, 0, 0, 0,
		date.Location(),
	)
	dayEnd := dayStart.AddDate(0, 0, 1)

	listDay := make([]storage.Event, 0, len(s.events))

	for _, e := range s.events {
		eventEndDate := e.Datetime.Add(e.Duration)

		if e.Datetime.Before(dayEnd) && eventEndDate.After(dayStart) {
			listDay = append(listDay, e)
		}
	}

	return listDay, nil
}

func (s *Storage) ListWeek(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_ = ctx
	weekStart := date.Truncate(24 * time.Hour)

	listWeek := make([]storage.Event, 0, len(s.events))

	for _, e := range s.events {
		eventEndDate := e.Datetime.Add(e.Duration)
		weekEnd := weekStart.AddDate(0, 0, 7)

		if e.Datetime.Before(weekEnd) && eventEndDate.After(weekStart) {
			listWeek = append(listWeek, e)
		}
	}

	return listWeek, nil
}

func (s *Storage) ListMonth(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_ = ctx
	monthStart := time.Date(
		date.Year(), date.Month(), 1,
		0, 0, 0, 0,
		date.Location(),
	)

	listMonth := make([]storage.Event, 0, len(s.events))

	for _, e := range s.events {
		eventEndDate := e.Datetime.Add(e.Duration)
		monthEnd := monthStart.AddDate(0, 1, 0)

		if e.Datetime.Before(monthEnd) && eventEndDate.After(monthStart) {
			listMonth = append(listMonth, e)
		}
	}

	return listMonth, nil
}
