package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/svetoslaven/tasktracker/internal/services"
)

func (app *application) handleJSONRequestBodyParseError(w http.ResponseWriter, r *http.Request, err error) {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var maxBytesError *http.MaxBytesError
	var timeParseError *time.ParseError

	var msg string

	switch {
	case errors.As(err, &syntaxError):
		msg = fmt.Sprintf("The body contains malformed JSON at character %d.", syntaxError.Offset)
	case errors.Is(err, io.ErrUnexpectedEOF):
		msg = "The body contains malformed JSON."
	case errors.As(err, &unmarshalTypeError):
		if unmarshalTypeError.Field != "" {
			msg = fmt.Sprintf("The body contains an incorrect JSON type for the field %q.", unmarshalTypeError.Field)
		} else {
			msg = fmt.Sprintf("The body contains an incorrect JSON type at character %d.", unmarshalTypeError.Offset)
		}
	case errors.Is(err, io.EOF):
		msg = "The body must not be empty."
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		msg = fmt.Sprintf("The body contains an unknown key %s.", fieldName)
	case errors.As(err, &maxBytesError):
		msg = fmt.Sprintf("The body must not be larger than %d bytes.", maxBytesError.Limit)
	case err.Error() == "expected EOF":
		msg = "The body must only contain a single JSON value."
	case strings.HasPrefix(err.Error(), "Time.UnmarshalJSON") || errors.As(err, &timeParseError):
		msg = "The body contains invalid time values."
	case errors.As(err, &invalidUnmarshalError):
		panic(err)
	default:
		msg = "The body could not be parsed properly."
	}

	app.sendErrorResponse(w, r, http.StatusBadRequest, msg)
}

func (app *application) handleServiceRetrievalError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	handler func(w http.ResponseWriter, r *http.Request),
) {
	switch {
	case errors.Is(err, services.ErrNoRecordsFound):
		handler(w, r)
	default:
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleServiceUpdateError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	handler func(w http.ResponseWriter, r *http.Request),
) {
	switch {
	case errors.Is(err, services.ErrEditConflict):
		handler(w, r)
	default:
		app.sendServerErrorResponse(w, r, err)
	}
}
