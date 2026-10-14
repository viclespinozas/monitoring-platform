-- +goose Up
CREATE TABLE services (
    id                     uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   text        NOT NULL CHECK (name <> ''),
    team                   text        NOT NULL CHECK (team <> ''),
    description            text        NOT NULL DEFAULT '',
    tags                   text[]      NOT NULL DEFAULT '{}',
    health_url             text        NOT NULL CHECK (health_url <> ''),
    check_interval_seconds integer     NOT NULL DEFAULT 30 CHECK (check_interval_seconds >= 5),
    timeout_seconds        integer     NOT NULL DEFAULT 5 CHECK (timeout_seconds > 0),
    expected_status        integer     NOT NULL DEFAULT 200 CHECK (expected_status BETWEEN 100 AND 599),
    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    deleted_at             timestamptz,
    CHECK (timeout_seconds < check_interval_seconds)
);

-- Soft delete: a name only has to be unique among live services,
-- so a deleted service's name can be registered again.
CREATE UNIQUE INDEX services_name_active_key ON services (name) WHERE deleted_at IS NULL;

CREATE INDEX services_team_idx ON services (team) WHERE deleted_at IS NULL;
CREATE INDEX services_tags_idx ON services USING gin (tags);

-- +goose Down
DROP TABLE services;
