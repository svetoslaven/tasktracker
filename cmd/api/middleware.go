package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
)

type middleware func(http.Handler) http.Handler

func (app *application) newMiddlewareChain(middlewares ...middleware) middleware {
	return middleware(func(next http.Handler) http.Handler {
		for i := 0; i < len(middlewares); i++ {
			next = middlewares[i](next)
		}

		return next
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection:", "close")
				app.sendServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r := app.setRequestContextUser(r, models.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerFields := strings.Fields(authorizationHeader)
		if len(headerFields) != 2 || headerFields[0] != "Bearer" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.sendUnauthorizedResponse(w, r, "Invalid Authorization header.")
			return
		}

		tokenPlaintext := headerFields[1]

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		user, ok := app.getTokenRecipient(ctx, w, r, tokenPlaintext, models.TokenScopeAuthentication)
		if !ok {
			return
		}

		r = app.setRequestContextUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireVerifiedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getRequestContextUser(r)

		if !user.IsVerified {
			msg := "Your account must be verified to access this resource."
			app.sendErrorResponse(w, r, http.StatusForbidden, msg)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := app.getRequestContextUser(r)

		if user.IsAnonymous() {
			app.sendUnauthorizedResponse(w, r, "You must be authenticated to access this resource.")
			return
		}

		next.ServeHTTP(w, r)
	}
}
