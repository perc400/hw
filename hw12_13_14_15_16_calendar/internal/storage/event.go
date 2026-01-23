package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID                string        `db:"id"`
	Title             string        `db:"title"`
	Datetime          time.Time     `db:"datetime"`
	Duration          time.Duration `db:"duration"`
	Description       string        `db:"description"`
	UserID            uint64        `db:"user_id"`
	NotificationDelay time.Duration `db:"notification_delay"`
	NotifiedAt        *time.Time    `db:"notified_at"`
}

var (
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
	ErrDateBusy           = errors.New("date is busy")
)
