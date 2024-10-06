package main

import (
	"context"
	"net/http"

	"github.com/svetoslaven/tasktracker/internal/models"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) setRequestContextUser(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) getRequestContextUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
