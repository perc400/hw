package app

import (
	"context"
)

type App struct { // TODO
}

type Logger interface { // TODO
}

type Storage interface { // TODO
}

//nolint:revive
func New(logger Logger, storage Storage) *App {
	return &App{}
}

//nolint:revive
func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
