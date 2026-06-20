CREATE TABLE tokens (
                        id         BIGSERIAL PRIMARY KEY,
                        user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        token_hash BYTEA NOT NULL,
                        scope      TEXT NOT NULL,
                        expires_at TIMESTAMP(0) WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX tokens_user_verification_unique
    ON tokens (user_id) WHERE scope = 'verification';