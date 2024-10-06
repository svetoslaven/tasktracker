package main

import (
	"context"
	"net/http"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
)

func (app *application) handleUserRegistration(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, validator, err := app.services.UserService.RegisterUser(ctx, input.Username, input.Email, input.Password)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	token, err := app.services.TokenService.GenerateToken(ctx, user.ID, 72*time.Hour, models.TokenScopeVerification)
	if err != nil {
		app.logError(err, r)
	} else {
		data := map[string]string{"verificationToken": token.Plaintext}
		app.sendEmail(user.Email, "user_welcome.tmpl", data)
	}

	if err := app.sendJSONResponse(w, http.StatusCreated, envelope{"user": user}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleUserVerification(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	recipient, ok := app.getTokenRecipient(ctx, w, r, input.TokenPlaintext, models.TokenScopeVerification)
	if !ok {
		return
	}

	err := app.services.TokenService.DeleteAllTokensForRecipient(ctx, recipient.ID, models.TokenScopeVerification)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	if err := app.services.UserService.VerifyUser(ctx, recipient); err != nil {
		app.handleServiceUpdateError(w, r, err, func(w http.ResponseWriter, r *http.Request) {
			app.sendEditConflictResponse(w, r)
		})
		return
	}

	msg := "You have successfully verified your account."
	if err := app.sendJSONResponse(w, http.StatusOK, app.newMessageEnvelope(msg), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleUserPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
		NewPassword    string `json:"new_password"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, ok := app.getTokenRecipient(ctx, w, r, input.TokenPlaintext, models.TokenScopePasswordReset)
	if !ok {
		return
	}

	err := app.services.TokenService.DeleteAllTokensForRecipient(ctx, user.ID, models.TokenScopePasswordReset)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	validator, err := app.services.UserService.ResetUserPassword(ctx, user, input.NewPassword)
	if err != nil {
		app.handleServiceUpdateError(w, r, err, func(w http.ResponseWriter, r *http.Request) {
			app.sendEditConflictResponse(w, r)
		})
		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	msg := "You have successfully reset your password."
	if err := app.sendJSONResponse(w, http.StatusOK, app.newMessageEnvelope(msg), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) getUserByEmail(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	email string,
	notFoundHandler func(w http.ResponseWriter, r *http.Request),
) (*models.User, bool) {
	user, validator, err := app.services.UserService.GetUserByEmail(ctx, email)
	if err != nil {
		app.handleServiceRetrievalError(w, r, err, notFoundHandler)
		return nil, false
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return nil, false
	}

	return user, true
}
