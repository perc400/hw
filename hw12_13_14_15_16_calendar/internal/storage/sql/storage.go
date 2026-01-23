package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/stdlib"                                 //nolint:revive,depguard,nolintlint
	"github.com/jackc/pgx/v5/pgconn"                                //nolint:depguard
	"github.com/jmoiron/sqlx"                                       //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type dbEvent struct {
	ID                string     `db:"id"`
	Title             string     `db:"title"`
	Datetime          time.Time  `db:"datetime"`
	Duration          int64      `db:"duration"`
	Description       string     `db:"description"`
	UserID            uint64     `db:"user_id"`
	NotificationDelay int64      `db:"notification_delay"`
	NotifiedAt        *time.Time `db:"notified_at"`
}

type Storage struct {
	db *sqlx.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func toDBEvent(e storage.Event) dbEvent {
	return dbEvent{
		ID:                e.ID,
		Title:             e.Title,
		Datetime:          e.Datetime,
		Duration:          int64(e.Duration.Seconds()),
		Description:       e.Description,
		UserID:            e.UserID,
		NotificationDelay: int64(e.NotificationDelay.Seconds()),
		NotifiedAt:        e.NotifiedAt,
	}
}

func toStorageEvents(dbEvents []dbEvent) []storage.Event {
	events := make([]storage.Event, 0, len(dbEvents))
	for _, dbEv := range dbEvents {
		events = append(events, storage.Event{
			ID:                dbEv.ID,
			Title:             dbEv.Title,
			Datetime:          dbEv.Datetime,
			Duration:          time.Duration(dbEv.Duration) * time.Second,
			Description:       dbEv.Description,
			UserID:            dbEv.UserID,
			NotificationDelay: time.Duration(dbEv.NotificationDelay) * time.Second,
			NotifiedAt:        dbEv.NotifiedAt,
		})
	}

	return events
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (s *Storage) Create(ctx context.Context, event storage.Event) error {
	const overlapQuery = `
	SELECT 1
	FROM events
	WHERE user_id = $1
		AND datetime < $2
		AND (datetime + duration * INTERVAL '1 second') > $3
	LIMIT 1;
	`

	var dummy int
	err := s.db.GetContext(
		ctx, &dummy, overlapQuery,
		event.UserID,
		event.Datetime.Add(event.Duration),
		event.Datetime,
	)
	if err == nil {
		return storage.ErrDateBusy
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	const insertQuery = `
	INSERT INTO events (
		id, title, datetime, duration, description, user_id, notification_delay
	)
	VALUES (
		:id, :title, :datetime, :duration, :description, :user_id, :notification_delay
	);
	`

	dbEv := toDBEvent(event)

	_, err = s.db.NamedExecContext(ctx, insertQuery, dbEv)
	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrEventAlreadyExists
		}
		return err
	}

	return nil
}

func (s *Storage) Update(ctx context.Context, eventID string, event storage.Event) error {
	const overlapQuery = `
	SELECT 1
	FROM events
	WHERE user_id = $1
		AND id != $2
		AND datetime < $3
		AND (datetime + duration * INTERVAL '1 second') > $4
	LIMIT 1;
	`

	var dummy int
	err := s.db.GetContext(
		ctx, &dummy, overlapQuery,
		event.UserID,
		eventID,
		event.Datetime.Add(event.Duration),
		event.Datetime,
	)
	if err == nil {
		return storage.ErrDateBusy
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	event.ID = eventID
	const updateQuery = `
	UPDATE events
	SET
		title = :title,
		datetime = :datetime,
		duration = :duration,
		description = :description,
		user_id = :user_id,
		notification_delay = :notification_delay
	WHERE id = :id;
	`

	dbEv := toDBEvent(event)

	res, err := s.db.NamedExecContext(ctx, updateQuery, dbEv)
	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) Delete(ctx context.Context, userID uint64, eventID string) error {
	const deleteQuery = `
	DELETE FROM events WHERE id = $1 AND user_id = $2;
	`

	res, err := s.db.ExecContext(ctx, deleteQuery, eventID, userID)
	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) ListDay(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	dayStart := time.Date(
		date.Year(), date.Month(), date.Day(),
		0, 0, 0, 0,
		date.Location(),
	)
	dayEnd := dayStart.AddDate(0, 0, 1)

	const query = `
	SELECT *
	FROM events
	WHERE user_id = $1 AND datetime < $2
		AND (datetime + duration * INTERVAL '1 second') > $3
	ORDER BY datetime;
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, userID, dayEnd, dayStart)
	if err != nil {
		return nil, err
	}

	return toStorageEvents(dbEvents), nil
}

func (s *Storage) ListWeek(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	weekStart := date.Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	const query = `
	SELECT *
	FROM events
	WHERE user_id = $1 AND datetime < $2
		AND (datetime + duration * INTERVAL '1 second') > $3
	ORDER BY datetime;
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, userID, weekEnd, weekStart)
	if err != nil {
		return nil, err
	}

	return toStorageEvents(dbEvents), nil
}

func (s *Storage) ListMonth(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	monthStart := time.Date(
		date.Year(), date.Month(), 1,
		0, 0, 0, 0,
		date.Location(),
	)
	monthEnd := monthStart.AddDate(0, 1, 0)

	const query = `
	SELECT *
	FROM events
	WHERE user_id = $1 AND datetime < $2
		AND (datetime + duration * INTERVAL '1 second') > $3
	ORDER BY datetime;
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, userID, monthEnd, monthStart)
	if err != nil {
		return nil, err
	}

	return toStorageEvents(dbEvents), nil
}

func (s *Storage) MarkNotified(ctx context.Context, eventID string, now time.Time) error {
	const query = `
	UPDATE events
	SET notified_at = $1
	WHERE id = $2 AND (notified_at IS NULL)
	`

	res, err := s.db.ExecContext(ctx, query, now, eventID)
	if err != nil {
		return err
	}

	affectedRows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affectedRows == 0 {
		return nil
	}

	return nil
}

func (s *Storage) ListEventsToNotify(ctx context.Context, now time.Time) ([]storage.Event, error) {
	var dbEvents []dbEvent

	const query = `
	SELECT *
	FROM events
	WHERE notification_delay > 0
		AND notified_at IS NULL
		AND datetime - (notification_delay * interval '1 second') <= $1
		AND $1 < datetime
	ORDER BY datetime;
	`

	err := s.db.SelectContext(ctx, &dbEvents, query, now)
	if err != nil {
		return nil, err
	}

	return toStorageEvents(dbEvents), nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, before time.Time) error {
	const deleteQuery = `
	DELETE FROM events WHERE datetime + (duration * INTERVAL '1 second') <= $1;
	`

	_, err := s.db.ExecContext(ctx, deleteQuery, before)
	if err != nil {
		return err
	}

	return nil
}
