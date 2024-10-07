package repositories

import (
	"context"
	"errors"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
)

var (
	ErrNoRecordsFound = errors.New("repositories: no matching records found")

	ErrEditConflict = errors.New("repositories: edit conflict")

	ErrDuplicateUsername = errors.New("repositories: duplicate username")
	ErrDuplicateEmail    = errors.New("repositories: duplicate email")

	ErrDuplicateTeamName = errors.New("repositories: duplicate team name")

	ErrInvitationExists = errors.New("repositories: invitation already exists")
)

type UserRepository interface {
	Insert(ctx context.Context, user *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

type TokenRepository interface {
	Insert(ctx context.Context, token *models.Token) error
	GetRecipient(ctx context.Context, tokenHash []byte, scope models.TokenScope) (*models.User, error)
	DeleteAllForRecipient(ctx context.Context, recipientID int64, scope models.TokenScope) error
}

type TeamRepository interface {
	InsertTeam(ctx context.Context, team *models.Team, creatorID int64) error
	GetTeamByName(ctx context.Context, name string, retrieverID int64) (*models.Team, error)
	GetAllTeams(ctx context.Context, filters models.TeamFilters, paginationOpts pagination.Options, retrieverID int64) ([]*models.Team, pagination.Metadata, error)
	GetMemberRole(ctx context.Context, teamID, memberID int64) (models.MemberRole, error)
	UpdateTeam(ctx context.Context, team *models.Team) error
	DeleteTeam(ctx context.Context, teamID int64) error

	InsertInvitation(ctx context.Context, teamID, inviterID, inviteeID int64) error
	GetAllInvitations(ctx context.Context, filters models.InvitationFilters, paginationOpts pagination.Options, retrieverID int64) ([]*models.Invitation, pagination.Metadata, error)
	AcceptInvitation(ctx context.Context, invitationID, inviteeID int64) error
	RejectInvitation(ctx context.Context, invitationID, inviteeID int64) error
	DeleteInvitation(ctx context.Context, invitationID, removerID int64) error

	GetMembership(ctx context.Context, teamID, memberID int64) (*models.Membership, error)
	GetAllTeamMembers(ctx context.Context, filters models.MembershipFilters, paginationOpts pagination.Options, teamID int64) ([]*models.Membership, pagination.Metadata, error)
	UpdateMembership(ctx context.Context, membership *models.Membership) error
	DeleteMembership(ctx context.Context, teamID, memberID int64) error
}

type RepositoryRegistry struct {
	UserRepo  UserRepository
	TokenRepo TokenRepository
	TeamRepo  TeamRepository
}
