CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    registered_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    username text UNIQUE NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    is_verified bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);