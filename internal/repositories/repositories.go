package repositories

import (
	"context"
	"errors"

	"github.com/svetoslaven/tasktracker/internal/models"
)

var (
	ErrDuplicateUsername = errors.New("repositories: duplicate username")
	ErrDuplicateEmail    = errors.New("repositories: duplicate email")
)

type UserRepository interface {
	Insert(ctx context.Context, user *models.User) error
}

type RepositoryRegistry struct {
	UserRepo UserRepository
}
