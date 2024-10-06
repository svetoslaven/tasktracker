package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/services"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

func (app *application) handleTeamCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		IsPublic bool   `json:"is_public"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	creator := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, validator, err := app.services.TeamService.CreateTeam(ctx, input.Name, input.IsPublic, creator.ID)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	if err := app.sendJSONResponse(w, http.StatusCreated, app.newTeamEnvelope(team), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTeamRetrievalByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("team_name")

	retriever := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, name, retriever.ID)
	if !ok {
		return
	}

	if err := app.sendJSONResponse(w, http.StatusOK, app.newTeamEnvelope(team), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleRetrievalOfAllTeams(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	validator := validator.New()

	var filters models.TeamFilters

	filters.Name = app.parseStringQueryParam(queryParams, "name", "")

	if queryParams.Has("is_public") {
		isPublic := app.parseBoolQueryParam(queryParams, "is_public", true, validator)
		filters.IsPublic = &isPublic
	}

	if validator.HasErrors() {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	paginationOpts := app.parsePaginationOptsFromQueryParams(queryParams, "name", []string{"name"}, validator)

	if validator.HasErrors() {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	retriever := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	teams, metadata, err := app.services.TeamService.GetAllTeams(ctx, filters, paginationOpts, retriever.ID)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	if err := app.sendJSONResponse(w, http.StatusOK, envelope{"teams": teams, "metadata": metadata}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTeamPartialUpdate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     *string `json:"name"`
		IsPublic *bool   `json:"is_public"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	teamName := r.PathValue("team_name")

	updater := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, teamName, updater.ID)
	if !ok {
		return
	}

	validator, err := app.services.TeamService.UpdateTeam(ctx, input.Name, input.IsPublic, team, updater.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEditConflict):
			app.sendEditConflictResponse(w, r)
		case errors.Is(err, services.ErrNoPermission):
			app.sendForbiddenResponse(w, r, "You do not have permission to edit this team.")
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	if err := app.sendJSONResponse(w, http.StatusOK, app.newTeamEnvelope(team), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTeamDeletion(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("team_name")

	remover := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, name, remover.ID)
	if !ok {
		return
	}

	if err := app.services.TeamService.DeleteTeam(ctx, team.ID, remover.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrNoRecordsFound):
			app.sendTeamNotFoundResponse(w, r)
		case errors.Is(err, services.ErrNoPermission):
			app.sendForbiddenResponse(w, r, "You do not have permission to delete this team.")
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	msg := "The team has been deleted successfully."
	if err := app.sendJSONResponse(w, http.StatusOK, app.newMessageEnvelope(msg), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) newTeamEnvelope(team *models.Team) envelope {
	return envelope{"team": team}
}

func (app *application) getTeamByName(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	name string,
	retrieverID int64,
) (*models.Team, bool) {
	team, err := app.services.TeamService.GetTeamByName(ctx, name, retrieverID)
	if err != nil {
		app.handleServiceRetrievalError(w, r, err, func(w http.ResponseWriter, r *http.Request) {
			app.sendTeamNotFoundResponse(w, r)
		})
		return nil, false
	}

	return team, true
}

func (app *application) sendTeamNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.sendNotFoundResponse(w, r, "A team with this name does not exist or you do not have permission to access it.")
}
