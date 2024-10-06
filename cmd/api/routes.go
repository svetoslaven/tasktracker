package main

import "net/http"

func (app *application) registerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", app.handleHealthCheck)

	mux.HandleFunc("POST /api/v1/tokens/verification", app.handleVerificationTokenCreation)

	mux.HandleFunc("POST /api/v1/users", app.handleUserRegistration)
	mux.HandleFunc("PUT /api/v1/users/verified", app.handleUserVerification)

	standardMiddlewareChain := app.newMiddlewareChain(app.recoverPanic)

	return standardMiddlewareChain(mux)
}
