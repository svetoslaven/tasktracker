package domain

import (
	"context"
	"fmt"
	"time"

	timefacade "github.com/svetoslaven/tasktracker/internal/facades/time"
	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/services"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

type TaskService struct {
	TaskRepo repositories.TaskRepository
	TeamRepo repositories.TeamRepository
}

func (s *TaskService) CreateTask(
	ctx context.Context,
	due time.Time,
	title, description string,
	priority string,
	creator, assignee *models.User,
	teamID int64,
) (*models.Task, *validator.Validator, error) {
	validator := validator.New()

	validator.Check(due.After(timefacade.Instance().Now()), "due", "Must be after the time of creation.")

	validator.CheckNonZero(title, "title")
	validator.CheckNonZero(description, "description")

	taskPriority, err := models.NewTaskPriority(priority)
	if err != nil {
		validator.AddError("priority", "Must be a valid task priority.")
	}

	if validator.HasErrors() {
		return nil, validator, nil
	}

	task := &models.Task{
		Due:         due,
		Title:       title,
		Description: description,
		Status:      models.TaskStatusOpen,
		Priority:    taskPriority,
		Creator:     creator,
		Assignee:    assignee,
	}

	creatorRole, err := s.TeamRepo.GetMemberRole(ctx, teamID, creator.ID)
	if err != nil {
		return nil, nil, err
	}

	if creatorRole < models.MemberRoleLeader {
		return nil, nil, services.ErrNoPermission
	}

	if err := s.TaskRepo.Insert(ctx, task, creator.ID, assignee.ID, teamID); err != nil {
		return nil, nil, err
	}

	return task, nil, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID, teamID int64) (*models.Task, error) {
	task, err := s.TaskRepo.GetByID(ctx, taskID, teamID)
	if err != nil {
		return nil, handleRepositoryRetrievalError(err)
	}

	return task, nil
}

func (s *TaskService) GetAllTasks(
	ctx context.Context,
	filters models.TaskFilters,
	status, priority []string,
	paginationOpts pagination.Options,
	teamID int64,
) ([]*models.Task, pagination.Metadata, *validator.Validator, error) {
	validator := validator.New()

	for _, s := range status {
		taskStatus, err := models.NewTaskStatus(s)
		if err != nil {
			validator.AddError("status", fmt.Sprintf("Contains an invalid task status %q", s))
			break
		}

		filters.Status = append(filters.Status, taskStatus)
	}

	for _, p := range priority {
		taskPriority, err := models.NewTaskPriority(p)
		if err != nil {
			validator.AddError("priority", fmt.Sprintf("Contains an invalid task priority %q", p))
			break
		}

		filters.Priority = append(filters.Priority, taskPriority)
	}

	if validator.HasErrors() {
		return nil, pagination.Metadata{}, validator, nil
	}

	tasks, metadata, err := s.TaskRepo.GetAll(ctx, filters, teamID, paginationOpts)
	return tasks, metadata, nil, err
}

func (s *TaskService) UpdateTaskStatus(
	ctx context.Context,
	task *models.Task,
	newStatus models.TaskStatus,
	updaterID int64,
) error {
	switch newStatus {
	case models.TaskStatusInProgress:
		return s.startTask(ctx, task, updaterID)
	case models.TaskStatusCompleted:
		return s.completeTask(ctx, task, updaterID)
	case models.TaskStatusCancelled:
		return s.cancelTask(ctx, task, updaterID)
	default:
		panic("invalid task status for update")
	}
}

func (s *TaskService) startTask(ctx context.Context, task *models.Task, updaterID int64) error {
	if updaterID != task.Assignee.ID {
		return services.ErrNoPermission
	}

	if task.Due.Before(timefacade.Instance().Now()) {
		return services.ErrTaskOverdue
	}

	if task.Status != models.TaskStatusOpen {
		return services.ErrTaskStatusConflict
	}

	if err := s.TaskRepo.UpdateTaskStatus(ctx, task, models.TaskStatusInProgress); err != nil {
		return handleRepositoryUpdateError(err)
	}

	return nil
}

func (s *TaskService) completeTask(ctx context.Context, task *models.Task, updaterID int64) error {
	if updaterID != task.Creator.ID {
		return services.ErrNoPermission
	}

	if task.Due.Before(timefacade.Instance().Now()) {
		return services.ErrTaskOverdue
	}

	if task.Status != models.TaskStatusInProgress {
		return services.ErrTaskStatusConflict
	}

	if err := s.TaskRepo.UpdateTaskStatus(ctx, task, models.TaskStatusCompleted); err != nil {
		return handleRepositoryUpdateError(err)
	}

	return nil
}

func (s *TaskService) cancelTask(ctx context.Context, task *models.Task, updaterID int64) error {
	if updaterID != task.Creator.ID {
		return services.ErrNoPermission
	}

	if task.Status != models.TaskStatusOpen && task.Status != models.TaskStatusInProgress {
		return services.ErrTaskStatusConflict
	}

	if err := s.TaskRepo.UpdateTaskStatus(ctx, task, models.TaskStatusCancelled); err != nil {
		return handleRepositoryUpdateError(err)
	}

	return nil
}
