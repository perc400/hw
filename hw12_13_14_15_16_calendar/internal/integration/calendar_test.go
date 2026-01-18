package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"        //nolint:depguard
	_ "github.com/jackc/pgx/stdlib" //nolint:depguard
	. "github.com/onsi/ginkgo/v2"   //nolint:revive,depguard
	. "github.com/onsi/gomega"      //nolint:revive
)

var _ = Describe("Calendar integration", func() {
	var (
		ctx context.Context
		db  *sql.DB
	)

	BeforeEach(func() {
		var err error

		ctx = context.Background()

		resp, err := http.Get("http://localhost:8080/health")
		Expect(err).NotTo(HaveOccurred())
		resp.Body.Close()

		db, err = sql.Open("pgx", "postgres://calendar_user:calendar@localhost:5432/calendar?sslmode=disable")
		Expect(err).NotTo(HaveOccurred())

		Expect(db.Ping()).To(Succeed())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	It("creates event and scheduler marks it notified", func() {
		start := time.Now().Add(10 * time.Second)
		eventID := uuid.NewString()

		resp, err := createEvent(ctx, 1, createEventRequest{
			ID:                eventID,
			Title:             "Test event",
			Datetime:          start.Format(time.RFC3339),
			Duration:          int64((10 * time.Minute).Seconds()),
			NotificationDelay: 1,
		})
		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		defer resp.Body.Close()

		By("waiting until scheduler marks event as notified")

		Eventually(func() bool {
			var notifiedAt sql.NullTime

			err := db.QueryRow(
				"select notified_at from events where id=$1",
				eventID,
			).Scan(&notifiedAt)

			return err == nil && notifiedAt.Valid
		}, 25*time.Second, 1*time.Second).Should(BeTrue())
	})

	It("prevents overlapping events for same user and allows for different users", func() {
		start := time.Now().Add(30 * time.Second)

		resp, err := createEvent(ctx, 2, createEventRequest{
			ID:                uuid.NewString(),
			Title:             "Main",
			Datetime:          start.Format(time.RFC3339),
			Duration:          int64((30 * time.Minute).Seconds()),
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		resp.Body.Close()

		resp2, err := createEvent(ctx, 2, createEventRequest{
			ID:                uuid.NewString(),
			Title:             "Overlap",
			Datetime:          start.Add(5 * time.Minute).Format(time.RFC3339),
			Duration:          int64((10 * time.Minute).Seconds()),
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp2.StatusCode).To(Equal(http.StatusConflict))
		resp2.Body.Close()

		resp3, err := createEvent(ctx, 3, createEventRequest{
			ID:                uuid.NewString(),
			Title:             "Other user",
			Datetime:          start.Add(5 * time.Minute).Format(time.RFC3339),
			Duration:          int64((10 * time.Minute).Seconds()),
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp3.StatusCode).To(Equal(http.StatusCreated))
		resp3.Body.Close()
	})

	It("lists events for the day", func() {
		start := time.Now().Add(1 * time.Minute)

		eventID := uuid.NewString()

		resp, err := createEvent(ctx, 4, createEventRequest{
			ID:                eventID,
			Title:             "List test",
			Datetime:          start.Format(time.RFC3339),
			Duration:          600,
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		resp.Body.Close()

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://localhost:8080/events/day?date="+time.Now().Format("2006-01-02"),
			nil,
		)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("X-User-ID", strconv.FormatUint(4, 10))

		resp, err = http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		defer resp.Body.Close()

		type listEvent struct {
			ID string `json:"id"`
		}
		var events []listEvent
		Expect(json.NewDecoder(resp.Body).Decode(&events)).To(Succeed())

		Expect(events).NotTo(BeEmpty())
	})

	It("updates event", func() {
		start := time.Now().Add(1 * time.Minute)
		eventID := uuid.NewString()

		resp, err := createEvent(ctx, 5, createEventRequest{
			ID:                eventID,
			Title:             "Old title",
			Datetime:          start.Format(time.RFC3339),
			Duration:          600,
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		resp.Body.Close()

		updatePayload := createEventRequest{
			ID:                eventID,
			Title:             "New title",
			Datetime:          start.Add(10 * time.Minute).Format(time.RFC3339),
			Duration:          1200,
			NotificationDelay: 1,
		}

		body, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodPut,
			"http://localhost:8080/events",
			bytes.NewReader(body),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", strconv.FormatUint(5, 10))

		resp, err = http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		resp.Body.Close()

		var title string
		err = db.QueryRow("select title from events where id=$1", eventID).Scan(&title)
		Expect(err).NotTo(HaveOccurred())
		Expect(title).To(Equal("New title"))
	})

	It("deletes event", func() {
		start := time.Now().Add(1 * time.Minute)
		eventID := uuid.NewString()

		resp, err := createEvent(ctx, 6, createEventRequest{
			ID:                eventID,
			Title:             "To delete",
			Datetime:          start.Format(time.RFC3339),
			Duration:          600,
			NotificationDelay: 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		resp.Body.Close()

		deletePayload := deleteEventRequest{
			ID: eventID,
		}

		body, _ := json.Marshal(deletePayload)

		req, _ := http.NewRequestWithContext(
			ctx,
			http.MethodDelete,
			"http://localhost:8080/events",
			bytes.NewReader(body),
		)
		req.Header.Set("X-User-ID", strconv.FormatUint(6, 10))

		resp, err = http.DefaultClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		resp.Body.Close()

		var count int
		err = db.QueryRow("select count(*) from events where id=$1", eventID).Scan(&count)
		Expect(err).NotTo(HaveOccurred())
		Expect(count).To(Equal(0))
	})
})
