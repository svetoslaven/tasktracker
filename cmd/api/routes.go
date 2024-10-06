package main

import "net/http"

func (app *application) registerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", app.handleHealthCheck)

	standardMiddlewareChain := app.newMiddlewareChain(app.recoverPanic)

	return standardMiddlewareChain(mux)
}
