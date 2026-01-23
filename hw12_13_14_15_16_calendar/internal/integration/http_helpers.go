package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type createEventRequest struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Datetime          string `json:"datetime"`
	Duration          int64  `json:"duration"`
	NotificationDelay int64  `json:"notification_delay"` //nolint:tagliatelle
}

type deleteEventRequest struct {
	ID string `json:"id"`
}

func createEvent(ctx context.Context, userID uint64, req createEventRequest) (*http.Response, error) {
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://localhost:8080/events",
		bytes.NewReader(body),
	)
	httpReq.Header.Set("X-User-ID", strconv.FormatUint(userID, 10))
	httpReq.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(httpReq)
}
