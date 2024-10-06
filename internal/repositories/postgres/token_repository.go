package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
)

type TokenRepository struct {
	DB *sql.DB
}

func (r *TokenRepository) Insert(ctx context.Context, token *models.Token) error {
	query := `
	INSERT INTO tokens (hash, recipient_id, expires_at, scope)
	VALUES ($1, $2, $3, $4)
	`

	args := []any{token.Hash, token.RecipientID, token.ExpiresAt, token.Scope}

	_, err := r.DB.ExecContext(ctx, query, args...)
	return err
}

func (r *TokenRepository) GetRecipient(
	ctx context.Context,
	tokenHash []byte,
	scope models.TokenScope,
) (*models.User, error) {
	query := `
	SELECT 
		users.id,
		users.username,
		users.email, 
		users.password_hash,
		users.is_verified,
		users.version
	FROM users
    INNER JOIN tokens
	ON users.id = tokens.recipient_id
    WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expires_at > $3
	`

	args := []any{tokenHash, scope, time.Now()}

	var user models.User

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(
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

func (r *TokenRepository) DeleteAllForRecipient(ctx context.Context, userID int64, scope models.TokenScope) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND recipient_id = $2
	`

	_, err := r.DB.ExecContext(ctx, query, scope, userID)
	return err
}
