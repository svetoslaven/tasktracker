package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/services"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

func (app *application) handleTaskCreation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Due              time.Time `json:"due"`
		Title            string    `json:"title"`
		Description      string    `json:"description"`
		Priority         string    `json:"priority"`
		AssigneeUsername string    `json:"assignee_username"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	creator := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), creator.ID)
	if !ok {
		return
	}

	assignee, ok := app.getUserByUsername(ctx, w, r, input.AssigneeUsername)
	if !ok {
		return
	}

	isAssigneeMember, err := app.services.TeamService.IsMember(ctx, team.ID, assignee.ID)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}

	if !isAssigneeMember {
		app.sendForbiddenResponse(w, r, "The assignee is not a member of this team.")
		return
	}

	task, validator, err := app.services.TaskService.CreateTask(
		ctx,
		input.Due,
		input.Title,
		input.Description,
		input.Priority,
		creator,
		assignee,
		team.ID,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoPermission):
			app.sendForbiddenResponse(w, r, "You do not have permission to assign tasks in this team.")
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	if err := app.sendJSONResponse(w, http.StatusCreated, app.newTaskEnvelope(task), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTaskRetrievalByID(w http.ResponseWriter, r *http.Request) {
	retriever := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), retriever.ID)
	if !ok {
		return
	}

	taskID, err := app.parseInt64PathParam(r, "task_id")
	if err != nil {
		app.sendTaskNotFoundResponse(w, r)
		return
	}

	task, ok := app.getTaskByID(ctx, w, r, taskID, team.ID)
	if !ok {
		return
	}

	if err := app.sendJSONResponse(w, http.StatusOK, app.newTaskEnvelope(task), nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleRetrievalOfAllTasks(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	validator := validator.New()

	var filters models.TaskFilters

	filters.CreatorUsername = app.parseStringQueryParam(queryParams, "creator_username", "")
	filters.AssigneeUsername = app.parseStringQueryParam(queryParams, "assignee_username", "")

	status := app.parseCSVQueryParam(queryParams, "status", []string{})
	priority := app.parseCSVQueryParam(queryParams, "priority", []string{})

	if queryParams.Has("created_before") {
		createdBefore := app.parseTimeQueryParam(queryParams, "created_before", time.Time{}, validator)
		filters.CreatedBefore = &createdBefore
	}

	if queryParams.Has("created_after") {
		createdAfter := app.parseTimeQueryParam(queryParams, "created_after", time.Time{}, validator)
		filters.CreatedAfter = &createdAfter
	}

	if queryParams.Has("due_before") {
		dueBefore := app.parseTimeQueryParam(queryParams, "due_before", time.Time{}, validator)
		filters.DueBefore = &dueBefore
	}

	if queryParams.Has("due_after") {
		dueAfter := app.parseTimeQueryParam(queryParams, "due_after", time.Time{}, validator)
		filters.DueAfter = &dueAfter
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

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), retriever.ID)
	if !ok {
		return
	}

	tasks, metadata, validator, err := app.services.TaskService.GetAllTasks(
		ctx,
		filters,
		status,
		priority,
		paginationOpts,
		team.ID,
	)
	if err != nil {
		app.sendServerErrorResponse(w, r, err)
		return
	}
	if validator != nil {
		app.sendValidationErrorResponse(w, r, validator.Errors)
		return
	}

	envelope := envelope{"tasks": tasks, "metadata": metadata}
	if err := app.sendJSONResponse(w, http.StatusOK, envelope, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTaskStart(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID int64 `json:"task_id"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	updater := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), updater.ID)
	if !ok {
		return
	}

	task, ok := app.getTaskByID(ctx, w, r, input.TaskID, team.ID)
	if !ok {
		return
	}

	err := app.services.TaskService.UpdateTaskStatus(ctx, task, models.TaskStatusInProgress, updater.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoPermission):
			app.sendForbiddenResponse(w, r, "Only the assignee can start the task.")
		case errors.Is(err, services.ErrTaskOverdue):
			app.sendOverdueTaskResponse(w, r)
		case errors.Is(err, services.ErrTaskStatusConflict):
			msg := fmt.Sprintf("Only tasks with %s status can be started", models.TaskStatusOpen.String())
			app.sendForbiddenResponse(w, r, msg)
		case errors.Is(err, services.ErrEditConflict):
			app.sendEditConflictResponse(w, r)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTaskCompletion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID int64 `json:"task_id"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	updater := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), updater.ID)
	if !ok {
		return
	}

	task, ok := app.getTaskByID(ctx, w, r, input.TaskID, team.ID)
	if !ok {
		return
	}

	err := app.services.TaskService.UpdateTaskStatus(ctx, task, models.TaskStatusCompleted, updater.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoPermission):
			msg := fmt.Sprintf("Only the task creator can mark the task as %s.", models.TaskStatusCompleted.String())
			app.sendForbiddenResponse(w, r, msg)
		case errors.Is(err, services.ErrTaskOverdue):
			app.sendOverdueTaskResponse(w, r)
		case errors.Is(err, services.ErrTaskStatusConflict):
			msg := fmt.Sprintf(
				"Only tasks with %s status can be marked as %s.",
				models.TaskStatusInProgress.String(), models.TaskStatusCompleted.String(),
			)
			app.sendForbiddenResponse(w, r, msg)
		case errors.Is(err, services.ErrEditConflict):
			app.sendEditConflictResponse(w, r)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) handleTaskCancellation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID int64 `json:"task_id"`
	}

	if err := app.parseJSONRequestBody(w, r, &input); err != nil {
		app.handleJSONRequestBodyParseError(w, r, err)
		return
	}

	updater := app.getRequestContextUser(r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	team, ok := app.getTeamByName(ctx, w, r, r.PathValue("team_name"), updater.ID)
	if !ok {
		return
	}

	task, ok := app.getTaskByID(ctx, w, r, input.TaskID, team.ID)
	if !ok {
		return
	}

	err := app.services.TaskService.UpdateTaskStatus(ctx, task, models.TaskStatusCancelled, updater.ID)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrNoPermission):
			msg := fmt.Sprintf("Only the task creator can mark the task as %s.", models.TaskStatusCancelled.String())
			app.sendForbiddenResponse(w, r, msg)
		case errors.Is(err, services.ErrTaskStatusConflict):
			msg := fmt.Sprintf(
				"Only tasks with %s an %s status can be marked as %s.",
				models.TaskStatusOpen.String(), models.TaskStatusInProgress.String(),
				models.TaskStatusCancelled.String(),
			)
			app.sendForbiddenResponse(w, r, msg)
		case errors.Is(err, services.ErrEditConflict):
			app.sendEditConflictResponse(w, r)
		default:
			app.sendServerErrorResponse(w, r, err)
		}

		return
	}

	if err := app.sendJSONResponse(w, http.StatusNoContent, envelope{}, nil); err != nil {
		app.sendServerErrorResponse(w, r, err)
	}
}

func (app *application) newTaskEnvelope(task *models.Task) envelope {
	return envelope{"task": task}
}

func (app *application) getTaskByID(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	taskID, teamID int64,
) (*models.Task, bool) {
	task, err := app.services.TaskService.GetTaskByID(ctx, taskID, teamID)
	if err != nil {
		app.handleServiceRetrievalError(w, r, err, app.sendTaskNotFoundResponse)
		return nil, false
	}

	return task, true
}

func (app *application) sendTaskNotFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.sendNotFoundResponse(w, r, "A task with this ID does not exist or it does not belong to this team.")
}

func (app *application) sendOverdueTaskResponse(w http.ResponseWriter, r *http.Request) {
	app.sendForbiddenResponse(w, r, "This task is overdue.")
}
