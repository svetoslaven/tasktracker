package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
)

func (app *application) handleVerificationTokenCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, ok := app.getUserByEmail(ctx, w, r, input.Email, func(w http.ResponseWriter, r *http.Request) {
		app.sendTokenEmailResponse(w, r, models.TokenScopeVerification)
	})
	if !ok {
		return
	}

	if user.IsVerified {
		app.sendErrorResponse(w, r, http.StatusUnprocessableEntity, "This user is already verified.")
		return
	}

	token, err := app.services.TokenService.GenerateToken(ctx, user.ID, 72*time.Hour, models.TokenScopeVerification)
	if err != nil {
		app.logError(err, r)
	} else {
		data := map[string]string{"verificationToken": token.Plaintext}
		app.sendEmail(user.Email, "verification_resend.tmpl", data)
	}

	app.sendTokenEmailResponse(w, r, models.TokenScopeVerification)
}

func (app *application) sendTokenEmailResponse(w http.ResponseWriter, r *http.Request, tokenScope models.TokenScope) {
	msg := fmt.Sprintf(
		"An email containing %s instructions will be sent if a user with the provided email address exists.",
		tokenScope.String(),
	)
	if err := app.sendJSONResponse(w, http.StatusAccepted, app.newMessageEnvelope(msg), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handlePasswordResetTokenCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, ok := app.getUserByEmail(ctx, w, r, input.Email, func(w http.ResponseWriter, r *http.Request) {
		app.sendTokenEmailResponse(w, r, models.TokenScopePasswordReset)
	})
	if !ok {
		return
	}

	token, err := app.services.TokenService.GenerateToken(ctx, user.ID, 24*time.Hour, models.TokenScopePasswordReset)
	if err != nil {
		app.logError(err, r)
	} else {
		data := map[string]string{"passwordResetToken": token.Plaintext}
		app.sendEmail(user.Email, "password_reset.tmpl", data)
	}

	app.sendTokenEmailResponse(w, r, models.TokenScopePasswordReset)
}

func (app *application) handleAuthenticationTokenCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, validator, err := app.services.UserService.GetUserByEmailAndPassword(ctx, input.Email, input.Password)
	if err != nil {
		app.handleServiceRetrievalError(w, r, err, func(w http.ResponseWriter, r *http.Request) {
			app.sendUnauthorizedResponse(w, r, "You have entered invalid credentials.")
		})

		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	token, err := app.services.TokenService.GenerateToken(ctx, user.ID, 24*time.Hour, models.TokenScopeAuthentication)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	if err := app.sendJSONResponse(w, http.StatusOK, envelope{"token": token.Plaintext}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}
}

func (app *application) getTokenRecipient(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	tokenPlaintext string,
	tokenScope models.TokenScope,
) (*models.User, bool) {
	recipient, validator, err := app.services.TokenService.GetTokenRecipient(ctx, tokenPlaintext, tokenScope)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return nil, false
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return nil, false
	}

	return recipient, true
}
