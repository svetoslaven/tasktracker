package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
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

func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID, memberID int64) (models.MemberRole, error) {
	query := `
	SELECT member_role
	FROM memberships
	WHERE team_id = $1 AND member_id = $2
	`

	var permissions models.MemberRole

	if err := r.DB.QueryRowContext(ctx, query, teamID, memberID).Scan(&permissions); err != nil {
		return models.MemberRoleRegular, handleQueryRowError(err)
	}

	return permissions, nil
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, team *models.Team) error {
	query := `
	UPDATE teams
	SET name = $1, is_public = $2, version = version + 1
	WHERE id = $3 AND version = $4
	RETURNING version
	`

	args := []any{team.Name, team.IsPublic, team.ID, team.Version}

	if err := r.DB.QueryRowContext(ctx, query, args...).Scan(&team.Version); err != nil {
		switch {
		case r.isDuplicateTeamNameError(err):
			return repositories.ErrDuplicateTeamName
		case errors.Is(err, sql.ErrNoRows):
			return repositories.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, teamID int64) error {
	query := `
	DELETE FROM teams
	WHERE id = $1 
	`

	err := delete(ctx, r.DB, query, teamID)
	return err
}

func (r *TeamRepository) InsertInvitation(ctx context.Context, teamID, inviterID, inviteeID int64) error {
	query := `
	INSERT INTO invitations (team_id, inviter_id, invitee_id)
	VALUES ($1, $2, $3)
	`

	if _, err := r.DB.ExecContext(ctx, query, teamID, inviterID, inviteeID); err != nil {
		switch {
		case r.isInvitationExistsError(err):
			return repositories.ErrInvitationExists
		default:
			return err
		}
	}

	return nil
}

func (r *TeamRepository) GetAllInvitations(
	ctx context.Context,
	filters models.InvitationFilters,
	paginationOpts pagination.Options,
	retrieverID int64,
) ([]*models.Invitation, pagination.Metadata, error) {
	var isRetrieverInviterCondition string

	if filters.IsInviter != nil {
		if *filters.IsInviter {
			isRetrieverInviterCondition = "AND invitations.inviter_id = $1"
		} else {
			isRetrieverInviterCondition = "AND invitations.inviter_id != $1"
		}
	}

	query := fmt.Sprintf(
		`
		SELECT
			count(*) OVER(),
			invitations.id,
			teams.name, teams.is_public,
			inviter.username, inviter.email, inviter.is_verified,
			invitee.username, invitee.email, invitee.is_verified
		FROM invitations
		INNER JOIN teams ON invitations.team_id = teams.id
		INNER JOIN users AS inviter ON inviter.id = invitations.inviter_id
		INNER JOIN users AS invitee ON invitee.id = invitations.invitee_id
		WHERE (inviter.id = $1 OR invitee.id = $1) AND teams.name ILIKE '%%' || $2 || '%%' %s
		ORDER BY id ASC
		LIMIT $3 OFFSET $4
		`,
		isRetrieverInviterCondition,
	)

	args := []any{retrieverID, filters.TeamName, paginationOpts.Limit(), paginationOpts.Offset()}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination.Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	invitations := []*models.Invitation{}

	for rows.Next() {
		var invitation models.Invitation
		invitation.Team = &models.Team{}
		invitation.Inviter = &models.User{}
		invitation.Invitee = &models.User{}

		err := rows.Scan(
			&totalRecords,
			&invitation.ID,
			&invitation.Team.Name, &invitation.Team.IsPublic,
			&invitation.Inviter.Username, &invitation.Inviter.Email, &invitation.Inviter.IsVerified,
			&invitation.Invitee.Username, &invitation.Invitee.Email, &invitation.Invitee.IsVerified,
		)
		if err != nil {
			return nil, pagination.Metadata{}, err
		}

		invitations = append(invitations, &invitation)
	}

	if err := rows.Err(); err != nil {
		return nil, pagination.Metadata{}, err
	}

	metadata := pagination.CalculateMetadata(paginationOpts.Page(), paginationOpts.PageSize(), totalRecords)
	return invitations, metadata, nil
}

func (r *TeamRepository) AcceptInvitation(ctx context.Context, invitationID, inviteeID int64) error {
	return runInTransaction(ctx, r.DB, nil, func(tx *sql.Tx) error {
		deleteInvitationQuery := `
		DELETE FROM invitations
		WHERE id = $1 AND invitee_id = $2
		RETURNING team_id
		`

		var teamID int64

		err := tx.QueryRowContext(ctx, deleteInvitationQuery, invitationID, inviteeID).Scan(&teamID)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return repositories.ErrNoRecordsFound
			default:
				return err
			}
		}

		if err := r.insertMembership(ctx, r.DB, teamID, inviteeID, models.MemberRoleRegular); err != nil {
			return err
		}

		return nil
	})
}

func (r *TeamRepository) RejectInvitation(ctx context.Context, invitationID, inviteeID int64) error {
	query := `
	DELETE FROM invitations
	WHERE id = $1 AND invitee_id = $2
	`

	err := delete(ctx, r.DB, query, invitationID, inviteeID)
	return err
}

func (r *TeamRepository) DeleteInvitation(ctx context.Context, invitationID, removerID int64) error {
	query := `
	DELETE FROM invitations
	WHERE id = $1 AND inviter_id = $2
	`

	err := delete(ctx, r.DB, query, invitationID, removerID)
	return err
}

func (r *TeamRepository) GetMembership(ctx context.Context, teamID, memberID int64) (*models.Membership, error) {
	query := `
	SELECT team_id, member_id, member_role, version
	FROM memberships
	WHERE team_id = $1 AND member_id = $2
	`

	var membership models.Membership
	membership.Member = &models.User{}

	err := r.DB.QueryRowContext(ctx, query, teamID, memberID).Scan(
		&membership.TeamID,
		&membership.Member.ID,
		&membership.MemberRole,
		&membership.Version,
	)
	if err != nil {
		return nil, handleQueryRowError(err)
	}

	return &membership, nil
}

func (r *TeamRepository) GetAllTeamMembers(
	ctx context.Context,
	filters models.MembershipFilters,
	paginationOpts pagination.Options,
	teamID int64,
) ([]*models.Membership, pagination.Metadata, error) {
	var filterByRoleCondition string

	args := []any{teamID, filters.MemberUsername, paginationOpts.Limit(), paginationOpts.Offset()}

	if len(filters.MemberRoles) > 0 {
		filterByRoleCondition = "AND memberships.member_role = ANY($5)"
		args = append(args, pq.Array(filters.MemberRoles))
	}

	query := fmt.Sprintf(
		`
		SELECT 
			count(*) OVER(),
			member.username, member.email, member.is_verified,
			memberships.member_role
		FROM memberships
		INNER JOIN users AS member ON member.id = memberships.member_id
		WHERE memberships.team_id = $1 AND member.username ILIKE '%%' || $2 || '%%' %s
		ORDER BY member.username ASC
		LIMIT $3 OFFSET $4
		`,
		filterByRoleCondition,
	)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination.Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	memberships := []*models.Membership{}

	for rows.Next() {
		var membership models.Membership
		membership.Member = &models.User{}

		err := rows.Scan(
			&totalRecords,
			&membership.Member.Username, &membership.Member.Email, &membership.Member.IsVerified,
			&membership.MemberRole,
		)
		if err != nil {
			return nil, pagination.Metadata{}, err
		}

		memberships = append(memberships, &membership)
	}

	if err := rows.Err(); err != nil {
		return nil, pagination.Metadata{}, err
	}

	metadata := pagination.CalculateMetadata(paginationOpts.Page(), paginationOpts.PageSize(), totalRecords)
	return memberships, metadata, nil
}

func (r *TeamRepository) UpdateMembership(ctx context.Context, membership *models.Membership) error {
	query := `
	UPDATE memberships
	SET member_role = $1, version = version + 1
	WHERE team_id = $2 AND member_id = $3 AND version = $4
	RETURNING version
	`

	args := []any{membership.MemberRole, membership.TeamID, membership.Member.ID, membership.Version}

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&membership.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return repositories.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (r *TeamRepository) DeleteMembership(ctx context.Context, teamID, memberID int64) error {
	query := `
	DELETE FROM memberships
	WHERE team_id = $1 AND member_id = $2
	`

	err := delete(ctx, r.DB, query, teamID, memberID)
	return err
}

func (r *TeamRepository) isDuplicateTeamNameError(err error) bool {
	return isDuplicateKeyError(err, "teams_name_key")
}

func (r *TeamRepository) isInvitationExistsError(err error) bool {
	return isDuplicateKeyError(err, "invitations_team_id_invitee_id_key")
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
