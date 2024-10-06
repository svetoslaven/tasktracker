package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/svetoslaven/tasktracker/internal/validator"
)

func (app *application) parseJSONRequestBody(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("expected EOF")
	}

	return nil
}

func (app *application) parseIntQueryParam(
	queryParams url.Values,
	key string,
	fallback int,
	validator *validator.Validator,
) int {
	value := queryParams.Get(key)

	if value == "" {
		return fallback
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		validator.AddError(key, "Must be an integer value.")
		return fallback
	}

	return intValue
}

func (app *application) parseStringQueryParam(queryParams url.Values, key, fallback string) string {
	value := queryParams.Get(key)

	if value == "" {
		return fallback
	}

	return value
}
