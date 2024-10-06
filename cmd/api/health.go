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
		app.sendServerErrorResponse(w, r, err)
		return
	}
}
