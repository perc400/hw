package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/queue" //nolint:depguard
)

type Sender struct {
	consumer queue.Consumer
	logger   Logger
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

func NewSender(consumer queue.Consumer, logg Logger) *Sender {
	return &Sender{
		consumer: consumer,
		logger:   logg,
	}
}

func (s *Sender) handleMessage(ctx context.Context, msg queue.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var n queue.Notification

	if err := json.Unmarshal(msg.Body, &n); err != nil {
		return err
	}

	s.logger.Info(
		fmt.Sprintf(
			"Send notification: user=%d, event=%s, title=%q, datetime=%s",
			n.UserID,
			n.EventID,
			n.Title,
			n.Datetime.Format(time.RFC3339),
		),
	)

	return nil
}

func (s *Sender) Start(ctx context.Context) error {
	return s.consumer.Consume(ctx, s.handleMessage)
}
