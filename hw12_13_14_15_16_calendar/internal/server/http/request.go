package internalhttp

import "time"

type CreateEventRequest struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Datetime          time.Time `json:"datetime"`
	Duration          int64     `json:"duration"`
	Description       string    `json:"description,omitempty"`
	NotificationDelay int64     `json:"notificationDelay,omitempty"`
}

type UpdateEventRequest struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Datetime          time.Time `json:"datetime"`
	Duration          int64     `json:"duration"`
	Description       string    `json:"description,omitempty"`
	NotificationDelay int64     `json:"notificationDelay,omitempty"`
}

type DeleteEventRequest struct {
	ID string `json:"id"`
}
