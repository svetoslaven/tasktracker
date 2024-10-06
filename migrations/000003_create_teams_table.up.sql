CREATE TABLE IF NOT EXISTS teams (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text UNIQUE NOT NULL,
    is_public bool NOT NULL,
    version integer NOT NULL DEFAULT 1
);