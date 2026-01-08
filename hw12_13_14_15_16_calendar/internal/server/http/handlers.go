package internalhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type Handlers struct {
	app    Application
	logger Logger
}

func NewHandler(app Application, logger Logger) *Handlers {
	return &Handlers{
		app:    app,
		logger: logger,
	}
}

func (h *Handlers) events(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createEvent(w, r)
	case http.MethodPut:
		h.updateEvent(w, r)
	case http.MethodDelete:
		h.deleteEvent(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func extractUserID(r *http.Request, logg Logger) (uint64, error) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		logg.Warn("missing X-User-ID header")
		return 0, errors.New("missing X-User-ID header")
	}
	return strconv.ParseUint(userID, 10, 64)
}

func (h *Handlers) createEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode create event request with error" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := storage.Event{
		ID:                req.ID,
		Title:             req.Title,
		Datetime:          req.Datetime,
		Duration:          time.Duration(req.Duration) * time.Second,
		Description:       req.Description,
		UserID:            userID,
		NotificationDelay: time.Duration(req.NotificationDelay) * time.Second,
	}

	err = h.app.CreateEvent(ctx, event)
	if err != nil {
		h.logger.Error("failed to create event" + err.Error())
		switch {
		case errors.Is(err, storage.ErrDateBusy):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, storage.ErrEventAlreadyExists):
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) updateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode update event request with error" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := storage.Event{
		ID:                req.ID,
		Title:             req.Title,
		Datetime:          req.Datetime,
		Duration:          time.Duration(req.Duration) * time.Second,
		Description:       req.Description,
		UserID:            userID,
		NotificationDelay: time.Duration(req.NotificationDelay) * time.Second,
	}

	err = h.app.UpdateEvent(ctx, event.ID, event)
	if err != nil {
		h.logger.Error("failed to update event" + err.Error())
		switch {
		case errors.Is(err, storage.ErrDateBusy):
			w.WriteHeader(http.StatusConflict)
		case errors.Is(err, storage.ErrEventNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) deleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var req DeleteEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode delete event request with error" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.app.DeleteEvent(ctx, userID, req.ID)
	if err != nil {
		h.logger.Error("failed to delete event" + err.Error())
		if errors.Is(err, storage.ErrEventNotFound) {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//nolint:dupl
func (h *Handlers) listDayEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.logger.Warn("invalid date value " + dateStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Warn("failed to parse date with error " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := h.app.ListDay(ctx, userID, date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(events)
	if err != nil {
		h.logger.Warn("failed to marshal response with error " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.logger.Warn("failed to write response with error " + err.Error())
	}
}

//nolint:dupl
func (h *Handlers) listWeekEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.logger.Warn("invalid date value " + dateStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Warn("failed to parse date with error " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := h.app.ListWeek(ctx, userID, date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(events)
	if err != nil {
		h.logger.Warn("failed to marshal response with error " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.logger.Warn("failed to write response with error " + err.Error())
	}
}

//nolint:dupl
func (h *Handlers) listMonthEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, err := extractUserID(r, h.logger)
	if err != nil {
		h.logger.Warn("invalid X-User-ID value " + r.Header.Get("X-User-ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.logger.Warn("invalid date value " + dateStr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.logger.Warn("failed to parse date with error " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := h.app.ListMonth(ctx, userID, date)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(events)
	if err != nil {
		h.logger.Warn("failed to marshal response with error " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp)
	if err != nil {
		h.logger.Warn("failed to write response with error " + err.Error())
	}
}
