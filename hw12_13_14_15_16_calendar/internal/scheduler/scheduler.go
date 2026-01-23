package scheduler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/queue"   //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type Scheduler struct {
	app       Application
	publisher queue.Publisher
	logger    Logger
	interval  time.Duration
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

type Application interface {
	MarkNotified(ctx context.Context, eventID string, now time.Time) error
	ListEventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error)
	DeleteOldEvents(ctx context.Context, before time.Time) error
}

func NewScheduler(logg Logger, app Application, publisher queue.Publisher, interval time.Duration) *Scheduler {
	return &Scheduler{
		app:       app,
		publisher: publisher,
		logger:    logg,
		interval:  interval,
	}
}

func (s *Scheduler) process(ctx context.Context) {
	now := time.Now()

	events, err := s.app.ListEventsToNotify(ctx, now)
	if err != nil {
		s.logger.Error("failed to list events to notify" + err.Error())
		return
	}

	for _, e := range events {
		notification := queue.Notification{
			EventID:  e.ID,
			Title:    e.Title,
			Datetime: e.Datetime,
			UserID:   e.UserID,
		}

		body, err := json.Marshal(notification)
		if err != nil {
			s.logger.Error("failed to marshall notification: " + err.Error())
			continue
		}

		if err := s.publisher.Publish(ctx, queue.Message{Body: body}); err != nil {
			s.logger.Error("failed to publish: " + err.Error())
			continue
		}

		if err := s.app.MarkNotified(ctx, e.ID, time.Now()); err != nil {
			s.logger.Error("failed to mark event as notified: " + err.Error())
		}
	}

	oldBefore := now.AddDate(-1, 0, 0)
	if err := s.app.DeleteOldEvents(ctx, oldBefore); err != nil {
		s.logger.Error("failed to delete old events: " + err.Error())
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("scheduler started")
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.process(ctx)
		}
	}
}
