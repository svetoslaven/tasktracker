package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

func (app *application) parseBoolQueryParam(
	queryParams url.Values,
	key string,
	fallback bool,
	validator *validator.Validator,
) bool {
	value := queryParams.Get(key)

	if value == "" {
		return fallback
	}

	value = strings.ToLower(value)

	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		validator.AddError(key, "Must be true or false.")
		return fallback
	}
}

func (app *application) parseInt64PathParam(r *http.Request, key string) (int64, error) {
	intValue, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		return 0, errors.New("invalid int64 path parameter")
	}

	return intValue, nil
}
