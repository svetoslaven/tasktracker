package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"golang.org/x/time/rate"
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

func (app *application) enforceURILength(next http.Handler) http.Handler {
	maxURIBytes := 1_048_576

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.RequestURI) > maxURIBytes {
			app.sendErrorResponse(w, r, http.StatusRequestURITooLong, "URI too long.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		clientLock sync.Mutex
		clients    = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			clientLock.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 1*time.Minute {
					delete(clients, ip)
				}
			}

			clientLock.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.cfg.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.sendServerErrorResponse(w, r, err)
				return
			}

			clientLock.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				clientLock.Unlock()
				app.sendErrorResponse(w, r, http.StatusTooManyRequests, "Rate limit exceeded.")
				return
			}

			clientLock.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin == "" {
			next.ServeHTTP(w, r)
		}

		for i := range app.cfg.cors.trustedOrigins {
			if origin != app.cfg.cors.trustedOrigins[i] {
				continue
			}

			w.Header().Add("Access-Control-Allow-Origin", origin)

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.WriteHeader(http.StatusOK)
				return
			}

			break
		}

		next.ServeHTTP(w, r)
	})
}
