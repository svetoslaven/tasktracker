package postgres

import (
	"context"
	"database/sql"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
)

type TeamRepository struct {
	DB *sql.DB
}

func (r *TeamRepository) InsertTeam(ctx context.Context, team *models.Team, creatorID int64) error {
	return runInTransaction(ctx, r.DB, nil, func(tx *sql.Tx) error {
		query := `
		INSERT INTO teams (name, is_public)
		VALUES ($1, $2)
		RETURNING id, version
		`

		args := []any{team.Name, team.IsPublic}

		if err := tx.QueryRowContext(ctx, query, args...).Scan(&team.ID, &team.Version); err != nil {
			switch {
			case r.isDuplicateTeamNameError(err):
				return repositories.ErrDuplicateTeamName
			default:
				return err
			}
		}

		if err := r.insertMembership(ctx, tx, team.ID, creatorID, models.MemberRoleOwner); err != nil {
			return err
		}

		return nil
	})
}

func (r *TeamRepository) isDuplicateTeamNameError(err error) bool {
	return isDuplicateKeyError(err, "teams_name_key")
}

func (r *TeamRepository) insertMembership(
	ctx context.Context,
	db dbExecutor,
	teamID, userID int64,
	role models.MemberRole,
) error {
	query := `
	INSERT INTO memberships (team_id, member_id, member_role)
	VALUES ($1, $2, $3)
	`

	_, err := db.ExecContext(ctx, query, teamID, userID, role)
	return err
}
