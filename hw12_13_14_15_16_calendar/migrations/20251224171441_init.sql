-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id                 UUID PRIMARY KEY,
    title              TEXT NOT NULL,
    datetime           TIMESTAMPTZ NOT NULL,
    duration           BIGINT NOT NULL,
    description        TEXT,
    user_id            BIGINT NOT NULL,
    notification_delay BIGINT
);

create index idx_events_user_id on events (user_id);
create index idx_events_datetime on events (datetime);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
-- +goose StatementEnd
