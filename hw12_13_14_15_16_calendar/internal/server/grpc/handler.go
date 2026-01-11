package internalgrpc

import (
	"context"
	"errors"

	eventpb "github.com/perc400/hw/hw12_13_14_15_calendar/api/event" //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage"  //nolint:depguard
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	eventpb.UnimplementedEventServiceServer
	app    Application
	logger Logger
}

func NewHandler(app Application, logger Logger) *Handler {
	return &Handler{
		app:    app,
		logger: logger,
	}
}

func (h *Handler) CreateEvent(
	ctx context.Context,
	req *eventpb.CreateEventRequest,
) (*eventpb.CreateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	event := pbEventToStorage(req.Event)

	err := h.app.CreateEvent(ctx, event)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrDateBusy):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, storage.ErrEventAlreadyExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			h.logger.Error("CreateEvent failed: " + err.Error())
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &eventpb.CreateEventResponse{}, nil
}

func (h *Handler) UpdateEvent(
	ctx context.Context,
	req *eventpb.UpdateEventRequest,
) (*eventpb.UpdateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	event := pbEventToStorage(req.Event)

	err := h.app.UpdateEvent(ctx, event.ID, event)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrDateBusy):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		case errors.Is(err, storage.ErrEventNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			h.logger.Error("UpdateEvent failed: " + err.Error())
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &eventpb.UpdateEventResponse{}, nil
}

func (h *Handler) DeleteEvent(
	ctx context.Context,
	req *eventpb.DeleteEventRequest,
) (*eventpb.DeleteEventResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event id is required")
	}
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := h.app.DeleteEvent(ctx, req.UserId, req.EventId)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEventNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			h.logger.Error("DeleteEvent failed: " + err.Error())
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &eventpb.DeleteEventResponse{}, nil
}

func (h *Handler) ListDay(ctx context.Context, req *eventpb.ListDayRequest) (*eventpb.ListEventsResponse, error) {
	if req.Date == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	events, err := h.app.ListDay(ctx, req.UserId, req.Date.AsTime())
	if err != nil {
		h.logger.Error("ListDay events failed: " + err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	result := make([]*eventpb.Event, 0, len(events))
	for _, e := range events {
		result = append(result, storageEventToPB(e))
	}

	return &eventpb.ListEventsResponse{
		Events: result,
	}, nil
}

func (h *Handler) ListWeek(ctx context.Context, req *eventpb.ListWeekRequest) (*eventpb.ListEventsResponse, error) {
	if req.Date == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	events, err := h.app.ListWeek(ctx, req.UserId, req.Date.AsTime())
	if err != nil {
		h.logger.Error("ListDay events failed: " + err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	result := make([]*eventpb.Event, 0, len(events))
	for _, e := range events {
		result = append(result, storageEventToPB(e))
	}

	return &eventpb.ListEventsResponse{
		Events: result,
	}, nil
}

func (h *Handler) ListMonth(ctx context.Context, req *eventpb.ListMonthRequest) (*eventpb.ListEventsResponse, error) {
	if req.Date == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	events, err := h.app.ListMonth(ctx, req.UserId, req.Date.AsTime())
	if err != nil {
		h.logger.Error("ListDay events failed: " + err.Error())
		return nil, status.Error(codes.Internal, "internal error")
	}

	result := make([]*eventpb.Event, 0, len(events))
	for _, e := range events {
		result = append(result, storageEventToPB(e))
	}

	return &eventpb.ListEventsResponse{
		Events: result,
	}, nil
}
