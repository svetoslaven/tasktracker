CREATE TABLE IF NOT EXISTS memberships (
    team_id bigint NOT NULL REFERENCES teams ON DELETE CASCADE,
    member_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    member_role integer NOT NULL,
    version integer NOT NULL DEFAULT 1,
    PRIMARY KEY (team_id, member_id)
);