package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"                                //nolint:depguard
	"github.com/jmoiron/sqlx"                                       //nolint:depguard
	"github.com/perc400/hw/hw12_13_14_15_calendar/internal/storage" //nolint:depguard
)

type Storage struct {
	db *sqlx.DB
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = db.PingContext(pingCtx)
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

	_, err = s.db.NamedExecContext(ctx, insertQuery, event)
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

	res, err := s.db.NamedExecContext(ctx, updateQuery, event)
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
	var events []storage.Event

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

	err := s.db.SelectContext(ctx, &events, query, userID, dayEnd, dayStart)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Storage) ListWeek(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	var events []storage.Event

	weekStart := date.Truncate(24 * time.Hour)
	weekEnd := weekStart.AddDate(0, 0, 7)

	const query = `
	SELECT *
	FROM events
	WHERE user_id = $1 AND datetime < $2
		AND (datetime + duration * INTERVAL '1 second') > $3
	ORDER BY datetime;
	`

	err := s.db.SelectContext(ctx, &events, query, userID, weekEnd, weekStart)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *Storage) ListMonth(ctx context.Context, userID uint64, date time.Time) ([]storage.Event, error) {
	var events []storage.Event

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

	err := s.db.SelectContext(ctx, &events, query, userID, monthEnd, monthStart)
	if err != nil {
		return nil, err
	}

	return events, nil
}
