package main

import "net/http"

func (app *application) sendValidationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.sendErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) sendServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(err, r)

	msg := "The server encountered a problem and could not process your request. Please try again later."
	app.sendErrorResponse(w, r, http.StatusInternalServerError, msg)
}

func (app *application) sendErrorResponse(w http.ResponseWriter, r *http.Request, status int, details any) {
	if err := app.sendJSONResponse(w, status, envelope{"error": details}, nil); err != nil {
		app.logError(err, r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) logError(err error, r *http.Request) {
	app.logger.LogError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}
