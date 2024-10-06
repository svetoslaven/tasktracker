package main

import (
	"context"
	"net/http"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
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

func (app *application) newTeamEnvelope(team *models.Team) envelope {
	return envelope{"team": team}
}
