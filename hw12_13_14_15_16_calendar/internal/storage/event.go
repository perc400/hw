package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID                string
	Title             string
	Datetime          time.Time
	Duration          time.Duration
	Description       string
	UserID            uint64
	NotificationDelay time.Duration
}

var (
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
	ErrDateBusy           = errors.New("date is busy")
)
