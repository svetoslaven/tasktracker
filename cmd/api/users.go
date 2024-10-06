package main

import (
	"context"
	"net/http"
	"time"
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

	if err := app.sendJSONResponse(w, http.StatusCreated, envelope{"user": user}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}
