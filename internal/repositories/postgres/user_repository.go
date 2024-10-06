package postgres

import (
	"context"
	"database/sql"

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

func (r *UserRepository) isDuplicateUsernameError(err error) bool {
	return isDuplicateKeyError(err, "users_username_key")
}

func (r *UserRepository) isDuplicateEmailError(err error) bool {
	return isDuplicateKeyError(err, "users_email_key")
}
