package repositories

import (
	"context"
	"errors"

	"github.com/svetoslaven/tasktracker/internal/models"
)

var (
	ErrNoRecordsFound = errors.New("repositories: no matching records found")

	ErrEditConflict = errors.New("repositories: edit conflict")

	ErrDuplicateUsername = errors.New("repositories: duplicate username")
	ErrDuplicateEmail    = errors.New("repositories: duplicate email")

	ErrDuplicateTeamName = errors.New("repositories: duplicate team name")
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
}

type RepositoryRegistry struct {
	UserRepo  UserRepository
	TokenRepo TokenRepository
	TeamRepo  TeamRepository
}
