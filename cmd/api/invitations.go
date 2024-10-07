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

func (app *application) handleInvitationCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TeamName        string `json:"team_name"`
		InviteeUsername string `json:"invitee_username"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	inviter := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, input.TeamName, inviter.ID)
	if !ok {
		return
	}

	invitee, ok := app.getUserByUsername(ctx, w, r, input.InviteeUsername)
	if !ok {
		return
	}

	if !invitee.IsVerified {
		app.sendForbiddenResponse(w, r, "The invitee is not a verified user.")
		return
	}

	isInviteeMember, err := app.services.TeamService.IsMember(ctx, team.ID, invitee.ID)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	if isInviteeMember {
		app.sendErrorResponse(w, r, http.StatusUnprocessableEntity, "The invitee is already a member")
		return
	}

	if err := app.services.TeamService.InviteUser(ctx, team.ID, inviter.ID, invitee.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrInvitationExists):
			app.sendErrorResponse(w, r, http.StatusUnprocessableEntity, "This user is already invited.")
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	msg := "The invitation has been created successfully."
	if err := app.sendJSONResponse(w, http.StatusCreated, app.newMessageEnvelope(msg), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleRetrievalOfAllInvitations(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	validator := validator.New()

	var filters models.InvitationFilters

	filters.TeamName = app.parseStringQueryParam(queryParams, "team_name", "")

	if queryParams.Has("is_inviter") {
		isInviter := app.parseBoolQueryParam(queryParams, "is_inviter", true, validator)
		filters.IsInviter = &isInviter
	}

	if validator.HasErrors() {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	paginationOpts := app.parsePaginationOptsFromQueryParams(queryParams, "", []string{""}, validator)

	if validator.HasErrors() {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	retriever := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	invitations, metadata, err := app.services.TeamService.GetAllInvitations(
		ctx,
		filters,
		paginationOpts,
		retriever.ID,
	)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	envelope := envelope{"invitations": invitations, "metadata": metadata}
	if err := app.sendJSONResponse(w, http.StatusOK, envelope, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleInvitationAccepting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		InvitationID int64 `json:"invitation_id"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	invitee := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := app.services.TeamService.AcceptInvitation(ctx, input.InvitationID, invitee.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrNoRecordsFound):
			app.sendInvitationNotFoundResponse(w, r, false)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleInvitationRejecting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		InvitationID int64 `json:"invitation_id"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	invitee := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := app.services.TeamService.RejectInvitation(ctx, input.InvitationID, invitee.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrNoRecordsFound):
			app.sendInvitationNotFoundResponse(w, r, false)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleInvitationDeletion(w http.ResponseWriter, r *http.Request) {
	invitationID, err := app.parseInt64PathParam(r, "invitation_id")
	if err != nil {
		app.sendInvitationNotFoundResponse(w, r, true)
		return
	}

	remover := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := app.services.TeamService.DeleteInvitation(ctx, invitationID, remover.ID); err != nil {
		switch {
		case errors.Is(err, services.ErrNoRecordsFound):
			app.sendInvitationNotFoundResponse(w, r, true)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) sendInvitationNotFoundResponse(w http.ResponseWriter, r *http.Request, isInviter bool) {
	msg := "An invitation with this ID does not exist or you are not the "

	if isInviter {
		msg += "inviter."
	} else {
		msg += "invitee."
	}

	app.sendNotFoundResponse(w, r, msg)
}
