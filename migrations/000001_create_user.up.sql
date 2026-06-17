CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY       NOT NULL,
    first_name    TEXT                        NOT NULL,
    second_name   TEXT                        NOT NULL,
    email         CITEXT UNIQUE               NOT NULL,
    password_hash BYTEA                       NOT NULL,
    activated     BOOL                        NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version       INTEGER                     NOT NULL DEFAULT 1
);