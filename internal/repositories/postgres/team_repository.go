package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
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

func (r *TeamRepository) GetTeamByName(ctx context.Context, name string, retrieverID int64) (*models.Team, error) {
	query := `
	SELECT id, name, is_public, version
	FROM teams
	WHERE name = $1 AND (is_public = true OR EXISTS(SELECT 1 FROM memberships WHERE team_id = id AND member_id = $2))
	`

	var team models.Team

	err := r.DB.QueryRowContext(ctx, query, name, retrieverID).Scan(
		&team.ID,
		&team.Name,
		&team.IsPublic,
		&team.Version,
	)
	if err != nil {
		return nil, handleQueryRowError(err)
	}

	return &team, nil
}

func (r *TeamRepository) GetAllTeams(
	ctx context.Context,
	filters models.TeamFilters,
	paginationOpts pagination.Options,
	retrieverID int64,
) ([]*models.Team, pagination.Metadata, error) {
	var filterByVisibilityCondition string

	args := []any{retrieverID, filters.Name, paginationOpts.Limit(), paginationOpts.Offset()}

	if filters.IsPublic != nil {
		filterByVisibilityCondition = fmt.Sprintf("AND teams.is_public = $%d", len(args)+1)
		args = append(args, filters.IsPublic)
	}

	query := fmt.Sprintf(
		`
		SELECT count(*) OVER(), teams.name, teams.is_public
		FROM teams
		LEFT JOIN memberships on memberships.team_id = teams.id AND memberships.member_id = $1
		WHERE teams.name ILIKE '%%' || $2 || '%%' AND (teams.is_public = true OR memberships.member_id IS NOT NULL) %s
		ORDER BY %s %s, name ASC
		LIMIT $3 OFFSET $4
		`,
		filterByVisibilityCondition,
		paginationOpts.SortColumn(), CalculateSortDirection(paginationOpts),
	)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination.Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	teams := []*models.Team{}

	for rows.Next() {
		var team models.Team

		if err := rows.Scan(&totalRecords, &team.Name, &team.IsPublic); err != nil {
			return nil, pagination.Metadata{}, err
		}

		teams = append(teams, &team)
	}

	if err := rows.Err(); err != nil {
		return nil, pagination.Metadata{}, err
	}

	metadata := pagination.CalculateMetadata(paginationOpts.Page(), paginationOpts.PageSize(), totalRecords)
	return teams, metadata, nil
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
