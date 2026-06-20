CREATE TABLE IF NOT EXISTS tokens
(
    user_id    BIGINT UNIQUE               NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    code_hash  BYTEA                       NOT NULL,
    scope      TEXT                        NOT NULL,
    expires_at TIMESTAMP(0) WITH TIME ZONE NOT NULL
);