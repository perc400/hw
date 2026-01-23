package internalhttp

import "time"

type EventResponse struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Datetime          time.Time `json:"datetime"`
	Duration          int64     `json:"duration"`
	Description       string    `json:"description,omitempty"`
	UserID            uint64    `json:"user_id"`                      //nolint:tagliatelle
	NotificationDelay int64     `json:"notification_delay,omitempty"` //nolint:tagliatelle
}
