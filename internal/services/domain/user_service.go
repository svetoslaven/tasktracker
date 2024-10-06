package domain

import (
	"context"
	"errors"

	bcryptfacade "github.com/svetoslaven/tasktracker/internal/facades/golang.org/x/crypto/bcrypt"
	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

const (
	usernameField = "username"
	emailField    = "email"
	passwordField = "password"
)

type UserService struct {
	UserRepo repositories.UserRepository
}

func (s *UserService) RegisterUser(
	ctx context.Context,
	username, email, password string,
) (*models.User, *validator.Validator, error) {
	validator := validator.New()

	s.validateUsername(username, validator)
	s.validateEmail(email, validator)
	s.validatePassword(password, validator)

	if validator.HasErrors() {
		return nil, validator, nil
	}

	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.UserRepo.Insert(ctx, user); err != nil {
		switch {
		case errors.Is(err, repositories.ErrDuplicateUsername):
			s.addDuplicateUsernameError(validator)
			return nil, validator, nil
		case errors.Is(err, repositories.ErrDuplicateEmail):
			s.addDuplicateEmailError(validator)
			return nil, validator, nil
		default:
			return nil, nil, err
		}
	}

	return user, nil, nil
}

func (s *UserService) validateUsername(username string, validator *validator.Validator) {
	validator.CheckNonZero(username, usernameField)
	validator.CheckStringMaxLength(username, 32, usernameField)
	validator.Check(
		s.isValidUsername(username),
		usernameField,
		"Must contain only alphanumeric characters or single hyphens, and must not begin or end with a hyphen.",
	)
}

func (s *UserService) validateEmail(email string, validator *validator.Validator) {
	validator.CheckNonZero(email, emailField)
	validator.CheckValidEmail(email, emailField)
}

func (s *UserService) validatePassword(password string, validator *validator.Validator) {
	validator.CheckNonZero(password, passwordField)
	validator.CheckStringMinLength(password, 8, passwordField)
	validator.CheckStringMaxLength(password, 72, passwordField)
}

func (s *UserService) isValidUsername(username string) bool {
	if len(username) == 0 {
		return false
	}

	if username[0] == '-' || username[len(username)-1] == '-' {
		return false
	}

	for i := 0; i < len(username); i++ {
		c := username[i]

		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}

		if c != '-' {
			return false
		}

		if i > 0 && username[i-1] == '-' {
			return false
		}
	}

	return true
}

func (s *UserService) addDuplicateUsernameError(validator *validator.Validator) {
	validator.AddError(usernameField, "A user with this username already exists.")
}

func (s *UserService) addDuplicateEmailError(validator *validator.Validator) {
	validator.AddError(emailField, "A user with this email address already exists.")
}

func (s *UserService) hashPassword(password string) ([]byte, error) {
	return bcryptfacade.Instance().GenerateFromPassword([]byte(password), 12)
}
