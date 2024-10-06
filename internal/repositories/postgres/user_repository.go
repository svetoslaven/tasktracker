package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) Insert(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (username, email, password_hash, is_verified)
	VALUES ($1, $2, $3, $4)
	RETURNING id, version
	`

	args := []any{user.Username, user.Email, user.PasswordHash, user.IsVerified}

	if err := r.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Version); err != nil {
		switch {
		case r.isDuplicateUsernameError(err):
			return repositories.ErrDuplicateUsername
		case r.isDuplicateEmailError(err):
			return repositories.ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, is_verified, version
	FROM users
	WHERE email = $1
	`

	var user models.User

	err := r.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsVerified,
		&user.Version,
	)

	if err != nil {
		return nil, handleQueryRowError(err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
	UPDATE users
	SET username = $1, email = $2, password_hash = $3, is_verified = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version
	`

	args := []any{user.Username, user.Email, user.PasswordHash, user.IsVerified, user.ID, user.Version}

	if err := r.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version); err != nil {
		switch {
		case r.isDuplicateUsernameError(err):
			return repositories.ErrDuplicateUsername
		case r.isDuplicateEmailError(err):
			return repositories.ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return repositories.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (r *UserRepository) isDuplicateUsernameError(err error) bool {
	return isDuplicateKeyError(err, "users_username_key")
}

func (r *UserRepository) isDuplicateEmailError(err error) bool {
	return isDuplicateKeyError(err, "users_email_key")
}
