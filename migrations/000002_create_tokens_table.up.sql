CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    recipient_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at timestamp(0) with time zone NOT NULL,
    scope integer NOT NULL
);