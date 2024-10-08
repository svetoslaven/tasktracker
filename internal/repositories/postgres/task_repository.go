package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
	"github.com/svetoslaven/tasktracker/internal/repositories"
)

type TaskRepository struct {
	DB *sql.DB
}

func (r *TaskRepository) Insert(ctx context.Context, task *models.Task, creatorID, assigneeID, teamID int64) error {
	query := `
	INSERT INTO tasks (due, title, description, status, priority, creator_id, assignee_id, team_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id, created_at, version
	`

	args := []any{task.Due, task.Title, task.Description, task.Status, task.Priority, creatorID, assigneeID, teamID}

	err := r.DB.QueryRowContext(ctx, query, args...).Scan(&task.ID, &task.CreatedAt, &task.Version)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID, teamID int64) (*models.Task, error) {
	query := `
	SELECT
		tasks.id,
		tasks.created_at,
		tasks.due,
		tasks.title,
		tasks.description,
		tasks.status,
		tasks.priority,
		tasks.version,
		creator.id, creator.username, creator.email, creator.is_verified,
		assignee.id, assignee.username, assignee.email, assignee.is_verified
	FROM tasks
	INNER JOIN users AS creator ON creator.id = tasks.creator_id
	INNER JOIN users AS assignee ON assignee.id = tasks.assignee_id
	WHERE tasks.id = $1 AND tasks.team_id = $2
	`

	var task models.Task
	task.Creator = &models.User{}
	task.Assignee = &models.User{}

	err := r.DB.QueryRowContext(ctx, query, taskID, teamID).Scan(
		&task.ID,
		&task.CreatedAt,
		&task.Due,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.Version,
		&task.Creator.ID, &task.Creator.Username, &task.Creator.Email, &task.Creator.IsVerified,
		&task.Assignee.ID, &task.Assignee.Username, &task.Assignee.Email, &task.Assignee.IsVerified,
	)
	if err != nil {
		return nil, handleQueryRowError(err)
	}

	return &task, nil
}

func (r *TaskRepository) GetAll(
	ctx context.Context,
	filters models.TaskFilters,
	teamID int64,
	paginationOpts pagination.Options,
) ([]*models.Task, pagination.Metadata, error) {
	var (
		filterByCreatedBeforeCondition string
		filterByCreatedAfterCondition  string
		filterByDueBeforeCondition     string
		filterByDueAfterCondition      string
		filterByStatusCondition        string
		filterByPriorityCondition      string
	)

	args := []any{
		teamID,
		filters.CreatorUsername,
		filters.AssigneeUsername,
		paginationOpts.Limit(),
		paginationOpts.Offset(),
	}

	if filters.CreatedBefore != nil {
		filterByCreatedBeforeCondition = fmt.Sprintf("AND tasks.created_at <= $%d", len(args)+1)
		args = append(args, *filters.CreatedBefore)
	}

	if filters.CreatedAfter != nil {
		filterByCreatedAfterCondition = fmt.Sprintf("AND tasks.created_at >= $%d", len(args)+1)
		args = append(args, *filters.CreatedAfter)
	}

	if filters.DueBefore != nil {
		filterByDueBeforeCondition = fmt.Sprintf("AND tasks.due <= $%d", len(args)+1)
		args = append(args, *filters.DueBefore)
	}

	if filters.DueAfter != nil {
		filterByDueAfterCondition = fmt.Sprintf("AND tasks.due >= $%d", len(args)+1)
		args = append(args, *filters.DueAfter)
	}

	if len(filters.Status) > 0 {
		filterByStatusCondition = fmt.Sprintf("AND tasks.status = ANY($%d)", len(args)+1)
		args = append(args, pq.Array(filters.Status))
	}

	if len(filters.Priority) > 0 {
		filterByPriorityCondition = fmt.Sprintf("AND tasks.priority = ANY($%d)", len(args)+1)
		args = append(args, pq.Array(filters.Priority))
	}

	query := fmt.Sprintf(
		`
		SELECT
			count(*) OVER(),
			tasks.id,
			tasks.created_at,
			tasks.due,
			tasks.title,
			tasks.description,
			tasks.status,
			tasks.priority,
			tasks.version,
			creator.username, creator.email, creator.is_verified,
			assignee.username, assignee.email, assignee.is_verified
		FROM tasks
		INNER JOIN users AS creator ON creator.id = tasks.creator_id
		INNER JOIN users AS assignee ON assignee.id = tasks.assignee_id
		WHERE tasks.team_id = $1
			AND creator.username ILIKE '%%' || $2 || '%%'
			AND assignee.username ILIKE '%%' || $3 || '%%'
			%s %s
			%s %s
			%s
			%s
		ORDER BY tasks.id ASC
		LIMIT $4 OFFSET $5
		`,
		filterByCreatedBeforeCondition, filterByCreatedAfterCondition,
		filterByDueBeforeCondition, filterByDueAfterCondition,
		filterByStatusCondition,
		filterByPriorityCondition,
	)

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pagination.Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	tasks := []*models.Task{}

	for rows.Next() {
		var task models.Task
		task.Creator = &models.User{}
		task.Assignee = &models.User{}

		err := rows.Scan(
			&totalRecords,
			&task.ID,
			&task.CreatedAt,
			&task.Due,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.Version,
			&task.Creator.Username, &task.Creator.Email, &task.Creator.IsVerified,
			&task.Assignee.Username, &task.Assignee.Email, &task.Assignee.IsVerified,
		)

		if err != nil {
			return nil, pagination.Metadata{}, err
		}

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, pagination.Metadata{}, err
	}

	metadata := pagination.CalculateMetadata(paginationOpts.Page(), paginationOpts.PageSize(), totalRecords)
	return tasks, metadata, nil
}

func (r *TaskRepository) UpdateTaskStatus(ctx context.Context, task *models.Task, newStatus models.TaskStatus) error {
	query := `
	UPDATE tasks
	SET status = $1, version = version + 1
	WHERE id = $2 AND version = $3
	RETURNING version
	`

	if err := r.DB.QueryRowContext(ctx, query, newStatus, task.ID, task.Version).Scan(&task.Status); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return repositories.ErrEditConflict
		default:
			return err
		}
	}

	return nil

}
