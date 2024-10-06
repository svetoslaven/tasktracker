package main

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func (app *application) sendJSONResponse(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(jsonData, '\n'))

	return nil
}
