package internalgrpc

import (
	"time"

	eventpb "github.com/perc400/hw/hw12_13_14_15_calendar/api/event" //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage"  //nolint:depguard
	"google.golang.org/protobuf/types/known/timestamppb"
)

func pbEventToStorage(e *eventpb.Event) storage.Event {
	return storage.Event{
		ID:                e.Id,
		Title:             e.Title,
		Datetime:          e.Datetime.AsTime(),
		Duration:          time.Duration(e.Duration) * time.Second,
		Description:       e.Description,
		UserID:            e.UserId,
		NotificationDelay: time.Duration(e.NotificationDelay) * time.Second,
	}
}

func storageEventToPB(e storage.Event) *eventpb.Event {
	return &eventpb.Event{
		Id:                e.ID,
		Title:             e.Title,
		Datetime:          timestamppb.New(e.Datetime),
		Duration:          int64(e.Duration.Seconds()),
		Description:       e.Description,
		UserId:            e.UserID,
		NotificationDelay: int64(e.NotificationDelay.Seconds()),
	}
}
