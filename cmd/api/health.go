package main

import (
	"fmt"
	"net/http"
)

func (app *application) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintln(w, "environment:", app.cfg.environment)
	fmt.Fprintln(w, "version:", version)
}
