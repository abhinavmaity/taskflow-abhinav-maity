package tasks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ProjectExists(ctx context.Context, projectID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM projects WHERE id = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, projectID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check project exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) HasProjectAccess(ctx context.Context, projectID, userID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM projects p
			LEFT JOIN tasks t ON t.project_id = p.id
			WHERE p.id = $1
				AND (p.owner_id = $2 OR t.assignee_id = $2 OR t.created_by = $2)
		)
	`
	var exists bool
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check project access: %w", err)
	}
	return exists, nil
}

func (r *Repository) UserExists(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`
	var exists bool
	if err := r.db.QueryRow(ctx, query, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}

func (r *Repository) ListTasks(ctx context.Context, projectID string, filters TaskFilters, page, limit int) ([]Task, int, error) {
	where := []string{"project_id = $1"}
	args := []any{projectID}
	nextArg := 2

	if filters.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", nextArg))
		args = append(args, filters.Status)
		nextArg++
	}
	if filters.Assignee != "" {
		where = append(where, fmt.Sprintf("assignee_id = $%d", nextArg))
		args = append(args, filters.Assignee)
		nextArg++
	}

	whereClause := strings.Join(where, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM tasks WHERE %s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	offset := (page - 1) * limit
	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT
			id::text,
			title,
			description,
			status::text,
			priority::text,
			project_id::text,
			assignee_id::text,
			created_by::text,
			COALESCE(to_char(due_date, 'YYYY-MM-DD'), ''),
			created_at,
			updated_at
		FROM tasks
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, nextArg, nextArg+1)

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks rows: %w", err)
	}
	return tasks, total, nil
}

func (r *Repository) CreateTask(ctx context.Context, task Task) (Task, error) {
	var dueDate any
	if task.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *task.DueDate)
		if err != nil {
			return Task{}, err
		}
		dueDate = parsed
	}

	const query = `
		INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id::text, title, description, status::text, priority::text, project_id::text,
			assignee_id::text, created_by::text, COALESCE(to_char(due_date, 'YYYY-MM-DD'), ''),
			created_at, updated_at
	`

	var created Task
	err := r.db.QueryRow(
		ctx,
		query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.ProjectID,
		task.AssigneeID,
		task.CreatedBy,
		dueDate,
	).Scan(
		&created.ID,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.Priority,
		&created.ProjectID,
		&created.AssigneeID,
		&created.CreatedBy,
		&created.DueDate,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return created, nil
}

func (r *Repository) GetTaskPermissionContext(ctx context.Context, taskID string) (TaskPermissionContext, error) {
	const query = `
		SELECT t.id::text, t.project_id::text, t.created_by::text, p.owner_id::text
		FROM tasks t
		INNER JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`

	var data TaskPermissionContext
	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&data.TaskID,
		&data.ProjectID,
		&data.CreatedBy,
		&data.ProjectOwner,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TaskPermissionContext{}, ErrNotFound
		}
		return TaskPermissionContext{}, fmt.Errorf("get task permission context: %w", err)
	}
	return data, nil
}

func (r *Repository) UpdateTask(ctx context.Context, taskID string, req UpdateTaskRequest) (Task, error) {
	setClauses := make([]string, 0, 6)
	args := make([]any, 0, 8)
	nextArg := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", nextArg))
		args = append(args, strings.TrimSpace(*req.Title))
		nextArg++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", nextArg))
		args = append(args, strings.TrimSpace(*req.Description))
		nextArg++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", nextArg))
		args = append(args, *req.Status)
		nextArg++
	}
	if req.Priority != nil {
		setClauses = append(setClauses, fmt.Sprintf("priority = $%d", nextArg))
		args = append(args, *req.Priority)
		nextArg++
	}
	if req.AssigneeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("assignee_id = $%d", nextArg))
		trimmed := strings.TrimSpace(*req.AssigneeID)
		if trimmed == "" {
			args = append(args, nil)
		} else {
			args = append(args, trimmed)
		}
		nextArg++
	}
	if req.DueDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", nextArg))
		trimmed := strings.TrimSpace(*req.DueDate)
		if trimmed == "" {
			args = append(args, nil)
		} else {
			parsed, err := time.Parse("2006-01-02", trimmed)
			if err != nil {
				return Task{}, err
			}
			args = append(args, parsed)
		}
		nextArg++
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, taskID)
	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s
		WHERE id = $%d
		RETURNING
			id::text, title, description, status::text, priority::text, project_id::text,
			assignee_id::text, created_by::text, COALESCE(to_char(due_date, 'YYYY-MM-DD'), ''),
			created_at, updated_at
	`, strings.Join(setClauses, ", "), nextArg)

	var updated Task
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Description,
		&updated.Status,
		&updated.Priority,
		&updated.ProjectID,
		&updated.AssigneeID,
		&updated.CreatedBy,
		&updated.DueDate,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return updated, nil
}

func (r *Repository) DeleteTask(ctx context.Context, taskID string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanTask(row interface {
	Scan(dest ...any) error
}) (Task, error) {
	var task Task
	var description sql.NullString
	var assigneeID sql.NullString
	var dueDateRaw string

	if err := row.Scan(
		&task.ID,
		&task.Title,
		&description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&assigneeID,
		&task.CreatedBy,
		&dueDateRaw,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return Task{}, fmt.Errorf("scan task: %w", err)
	}

	if description.Valid {
		value := description.String
		task.Description = &value
	}
	if assigneeID.Valid {
		value := assigneeID.String
		task.AssigneeID = &value
	}
	if dueDateRaw != "" {
		value := dueDateRaw
		task.DueDate = &value
	}

	return task, nil
}
