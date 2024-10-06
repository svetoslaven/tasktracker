package main

import (
	"net/url"

	"github.com/svetoslaven/tasktracker/internal/pagination"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

func (app *application) parsePaginationOptsFromQueryParams(
	queryParams url.Values,
	defaultSort string,
	sortSafelist []string,
	validator *validator.Validator,
) pagination.Options {
	page := app.parseIntQueryParam(queryParams, pagination.PageKey, 1, validator)
	pageSize := app.parseIntQueryParam(queryParams, pagination.PageSizeKey, 20, validator)
	sort := app.parseStringQueryParam(queryParams, pagination.SortKey, defaultSort)

	if validator.HasErrors() {
		return pagination.Options{}
	}

	opts := pagination.NewOptions(page, pageSize, sort, sortSafelist)

	opts.Validate(validator)

	if validator.HasErrors() {
		return pagination.Options{}
	}

	return opts
}
