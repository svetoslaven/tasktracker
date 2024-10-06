package domain

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

const tokenField = "token"

type TokenService struct {
	TokenRepo repositories.TokenRepository
}

func (s *TokenService) GenerateToken(
	ctx context.Context,
	recipientID int64,
	ttl time.Duration,
	scope models.TokenScope,
) (*models.Token, error) {
	token := &models.Token{
		RecipientID: recipientID,
		ExpiresAt:   time.Now().Add(ttl),
		Scope:       scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	token.Hash = s.calculateTokenHash(token.Plaintext)

	if err := s.TokenRepo.Insert(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *TokenService) GetTokenRecipient(
	ctx context.Context,
	tokenPlaintext string,
	scope models.TokenScope,
) (*models.User, *validator.Validator, error) {
	validator := validator.New()

	validator.CheckNonZero(tokenPlaintext, tokenField)

	if validator.HasErrors() {
		return nil, validator, nil
	}

	tokenHash := s.calculateTokenHash(tokenPlaintext)

	user, err := s.TokenRepo.GetRecipient(ctx, tokenHash, scope)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			validator.AddError(tokenField, fmt.Sprintf("Invalid or expired %s token.", scope.String()))
			return nil, validator, nil
		default:
			return nil, nil, err
		}
	}

	return user, nil, nil
}

func (s *TokenService) DeleteAllTokensForRecipient(
	ctx context.Context,
	recipientID int64,
	scope models.TokenScope,
) error {
	return s.TokenRepo.DeleteAllForRecipient(ctx, recipientID, scope)
}

func (s *TokenService) calculateTokenHash(tokenPlaintext string) []byte {
	hash := sha256.Sum256([]byte(tokenPlaintext))
	return hash[:]
}
