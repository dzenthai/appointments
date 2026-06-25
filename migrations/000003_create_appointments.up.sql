CREATE TABLE IF NOT EXISTS appointments
(
    id          BIGSERIAL PRIMARY KEY                          NOT NULL,
    client_id   BIGINT REFERENCES users (id) ON DELETE CASCADE NOT NULL,
    provider_id BIGINT REFERENCES users (id) ON DELETE CASCADE NOT NULL CHECK (client_id <> provider_id),
    title       TEXT                                           NOT NULL,
    description TEXT,
    starts_at   TIMESTAMP(0) WITH TIME ZONE                    NOT NULL,
    ends_at     TIMESTAMP(0) WITH TIME ZONE                    NOT NULL CHECK (ends_at > starts_at),
    status      TEXT                                           NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'confirmed', 'cancelled', 'completed')),
    created_at  TIMESTAMP(0) WITH TIME ZONE                    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP(0) WITH TIME ZONE                    NOT NULL DEFAULT NOW(),
    version     BIGINT                                         NOT NULL DEFAULT 1
);