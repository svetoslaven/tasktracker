package services

import (
	"context"
	"errors"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

var (
	ErrNoRecordsFound = errors.New("services: no matching records found")

	ErrEditConflict = errors.New("services: edit conflict")

	ErrNoPermission = errors.New("services: no permission")

	ErrInvitationExists = errors.New("services: invitation already exists")
)

type UserService interface {
	RegisterUser(ctx context.Context, username, email, password string) (*models.User, *validator.Validator, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, *validator.Validator, error)
	GetUserByEmailAndPassword(ctx context.Context, email, password string) (*models.User, *validator.Validator, error)
	VerifyUser(ctx context.Context, user *models.User) error
	ResetUserPassword(ctx context.Context, user *models.User, newPassword string) (*validator.Validator, error)
}

type TokenService interface {
	GenerateToken(ctx context.Context, recipientID int64, ttl time.Duration, scope models.TokenScope) (*models.Token, error)
	GetTokenRecipient(ctx context.Context, tokenPlaintext string, scope models.TokenScope) (*models.User, *validator.Validator, error)
	DeleteAllTokensForRecipient(ctx context.Context, recipientID int64, scope models.TokenScope) error
}

type TeamService interface {
	CreateTeam(ctx context.Context, name string, isPublic bool, creatorID int64) (*models.Team, *validator.Validator, error)
	GetTeamByName(ctx context.Context, name string, retrieverID int64) (*models.Team, error)
	GetAllTeams(ctx context.Context, filters models.TeamFilters, paginationOpts pagination.Options, retrieverID int64) ([]*models.Team, pagination.Metadata, error)
	UpdateTeam(ctx context.Context, newName *string, newIsPublic *bool, team *models.Team, updaterID int64) (*validator.Validator, error)
	DeleteTeam(ctx context.Context, teamID, removerID int64) error

	IsMember(ctx context.Context, teamID, userID int64) (bool, error)
	InviteUser(ctx context.Context, teamID, inviterID, inviteeID int64) error
	GetAllInvitations(ctx context.Context, filters models.InvitationFilters, paginationOpts pagination.Options, retrieverID int64) ([]*models.Invitation, pagination.Metadata, error)
	AcceptInvitation(ctx context.Context, invitationID, inviteeID int64) error
	RejectInvitation(ctx context.Context, invitationID, inviteeID int64) error
	DeleteInvitation(ctx context.Context, invitationID, removerID int64) error
}

type ServiceRegistry struct {
	UserService  UserService
	TokenService TokenService
	TeamService  TeamService
}
