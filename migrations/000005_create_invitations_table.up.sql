CREATE TABLE IF NOT EXISTS invitations (
    id bigserial PRIMARY KEY,
    team_id bigint NOT NULL REFERENCES teams ON DELETE CASCADE,
    inviter_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    invitee_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    UNIQUE(team_id, invitee_id)
);