package main

import "net/http"

func (app *application) registerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", app.handleHealthCheck)

	mux.HandleFunc("POST /api/v1/tokens/verification", app.handleVerificationTokenCreation)
	mux.HandleFunc("POST /api/v1/tokens/password-reset", app.handlePasswordResetTokenCreation)
	mux.HandleFunc("POST /api/v1/tokens/authentication", app.handleAuthenticationTokenCreation)

	mux.HandleFunc("POST /api/v1/users", app.handleUserRegistration)
	mux.HandleFunc("GET /api/v1/users/{username}", app.requireVerifiedUser(app.handleUserRetrievalByUsername))
	mux.HandleFunc("PUT /api/v1/users/verified", app.handleUserVerification)
	mux.HandleFunc("PUT /api/v1/users/password", app.handleUserPasswordReset)

	mux.HandleFunc("POST /api/v1/teams", app.requireVerifiedUser(app.handleTeamCreation))
	mux.HandleFunc("GET /api/v1/teams/{team_name}", app.requireVerifiedUser(app.handleTeamRetrievalByName))
	mux.HandleFunc("GET /api/v1/teams", app.requireVerifiedUser(app.handleRetrievalOfAllTeams))
	mux.HandleFunc("PATCH /api/v1/teams/{team_name}", app.requireVerifiedUser(app.handleTeamPartialUpdate))
	mux.HandleFunc("DELETE /api/v1/teams/{team_name}", app.requireVerifiedUser(app.handleTeamDeletion))

	standardMiddlewareChain := app.newMiddlewareChain(app.recoverPanic, app.authenticate)

	return standardMiddlewareChain(mux)
}
