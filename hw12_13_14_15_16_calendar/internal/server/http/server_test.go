package internalhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/app"                          //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/logger"                       //nolint:depguard
	internalhttp "github.com/perc400/hw/hw12_13_14_15_calendar/internal/server/http"     //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage"                      //nolint:depguard
	memorystorage "github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage/memory" //nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestHTTPCreateAndListDay(t *testing.T) {
	logg, err := logger.New("Info")
	require.NoError(t, err)

	strg := memorystorage.New()
	a := app.New(logg, strg)

	handler := internalhttp.NewHandlerForTest(logg, a)
	server := httptest.NewServer(handler)
	defer server.Close()

	userID := uint64(42)
	date := time.Date(2025, 1, 8, 12, 0, 0, 0, time.UTC)

	createReq := map[string]any{
		"id":                 "event-1",
		"title":              "Test event",
		"datetime":           date,
		"duration":           int64(3600),
		"description":        "test",
		"notification_delay": int64(60),
	}

	body, err := json.Marshal(createReq)
	require.NoError(t, err)

	ctxTmPost, cancel := context.WithTimeout(
		context.Background(),
		time.Second*5,
	)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctxTmPost,
		http.MethodPost,
		server.URL+"/events",
		bytes.NewReader(body),
	)
	require.NoError(t, err)

	req.Header.Set("X-User-ID", strconv.FormatUint(userID, 10))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	ctxTmGet, cancel := context.WithTimeout(
		context.Background(),
		time.Second*5,
	)
	defer cancel()

	req, err = http.NewRequestWithContext(
		ctxTmGet,
		http.MethodGet,
		server.URL+"/events/day?date=2025-01-08",
		bytes.NewReader(body),
	)
	require.NoError(t, err)

	req.Header.Set("X-User-ID", strconv.FormatUint(userID, 10))

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result []storage.Event
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	require.Len(t, result, 1)
	require.Equal(t, "event-1", result[0].ID)
}
