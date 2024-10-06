package services

import (
	"context"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

type UserService interface {
	RegisterUser(ctx context.Context, username, email, password string) (*models.User, *validator.Validator, error)
}

type ServiceRegistry struct {
	UserService UserService
}
