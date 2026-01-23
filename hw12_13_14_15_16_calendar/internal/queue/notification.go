package queue

import "time"

type Notification struct {
	EventID  string    `json:"event_id"` //nolint:tagliatelle
	Title    string    `json:"title"`
	Datetime time.Time `json:"datetime"`
	UserID   uint64    `json:"user_id"` //nolint:tagliatelle
}
