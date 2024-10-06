package services

import (
	"context"
	"errors"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

var (
	ErrNoRecordsFound = errors.New("services: no matching records found")

	ErrEditConflict = errors.New("services: edit conflict")
)

type UserService interface {
	RegisterUser(ctx context.Context, username, email, password string) (*models.User, *validator.Validator, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, *validator.Validator, error)
	VerifyUser(ctx context.Context, user *models.User) error
	ResetUserPassword(ctx context.Context, user *models.User, newPassword string) (*validator.Validator, error)
}

type TokenService interface {
	GenerateToken(ctx context.Context, recipientID int64, ttl time.Duration, scope models.TokenScope) (*models.Token, error)
	GetTokenRecipient(ctx context.Context, tokenPlaintext string, scope models.TokenScope) (*models.User, *validator.Validator, error)
	DeleteAllTokensForRecipient(ctx context.Context, recipientID int64, scope models.TokenScope) error
}

type ServiceRegistry struct {
	UserService  UserService
	TokenService TokenService
}
