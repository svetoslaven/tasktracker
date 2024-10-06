package domain

import (
	"context"
	"errors"

	bcryptfacade "github.com/svetoslaven/tasktracker/internal/facades/golang.org/x/crypto/bcrypt"
	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/services"
	"github.com/svetoslaven/tasktracker/internal/validator"
	"golang.org/x/crypto/bcrypt"
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

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, *validator.Validator, error) {
	validator := validator.New()

	s.validateEmail(email, validator)

	if validator.HasErrors() {
		return nil, validator, nil
	}

	user, err := s.getByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	return user, nil, nil
}

func (s *UserService) GetUserByEmailAndPassword(
	ctx context.Context,
	email, password string,
) (*models.User, *validator.Validator, error) {
	validator := validator.New()

	s.validateEmail(email, validator)
	s.validatePassword(password, validator)

	if validator.HasErrors() {
		return nil, validator, nil
	}

	user, err := s.getByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	if err := bcryptfacade.Instance().CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return nil, nil, services.ErrNoRecordsFound
		default:
			return nil, nil, err
		}
	}

	return user, nil, nil
}

func (s *UserService) VerifyUser(ctx context.Context, user *models.User) error {
	user.IsVerified = true

	if err := s.UserRepo.Update(ctx, user); err != nil {
		return handleRepositoryUpdateError(err)
	}

	return nil
}

func (s *UserService) ResetUserPassword(
	ctx context.Context,
	user *models.User,
	newPassword string,
) (*validator.Validator, error) {
	validator := validator.New()

	s.validatePassword(newPassword, validator)

	if validator.HasErrors() {
		return validator, nil
	}

	passwordHash, err := s.hashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = passwordHash

	if err := s.UserRepo.Update(ctx, user); err != nil {
		return nil, handleRepositoryUpdateError(err)
	}

	return nil, nil
}

func (s *UserService) getByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, handleRepositoryRetrievalError(err)
	}

	return user, nil
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
