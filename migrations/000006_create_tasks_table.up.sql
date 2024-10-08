CREATE TABLE IF NOT EXISTS tasks (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    due timestamp(0) with time zone NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    status integer NOT NULL,
    priority integer NOT NULL,
    creator_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    assignee_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    team_id bigint NOT NULL REFERENCES teams ON DELETE CASCADE,
    version integer NOT NULL DEFAULT 1
);