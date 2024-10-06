package main

import (
	"net/http"
)

func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	envelope := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.cfg.environment,
			"version":     version,
		},
	}

	err := app.sendJSONResponse(w, http.StatusOK, envelope, nil)
	if err != nil {
		msg := "The server encountered a problem and could not process your request. Please try again later."
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
}
