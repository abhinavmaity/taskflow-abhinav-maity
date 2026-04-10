package projects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAccessibleProjects(ctx context.Context, userID string, page, limit int) ([]Project, int, error) {
	const countQuery = `
		SELECT COUNT(DISTINCT p.id)
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.created_by = $1
	`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count accessible projects: %w", err)
	}

	const listQuery = `
		SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id
		WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.created_by = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (page - 1) * limit
	rows, err := r.db.Query(ctx, listQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list accessible projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, 0, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate projects rows: %w", err)
	}

	return projects, total, nil
}

func (r *Repository) CreateProject(ctx context.Context, project Project) (Project, error) {
	const query = `
		INSERT INTO projects (id, name, description, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, owner_id, created_at
	`

	var created Project
	err := r.db.QueryRow(ctx, query, project.ID, project.Name, project.Description, project.OwnerID).Scan(
		&created.ID,
		&created.Name,
		&created.Description,
		&created.OwnerID,
		&created.CreatedAt,
	)
	if err != nil {
		return Project{}, fmt.Errorf("insert project: %w", err)
	}
	return created, nil
}

func (r *Repository) GetProjectByID(ctx context.Context, projectID string) (Project, error) {
	const query = `
		SELECT id, name, description, owner_id, created_at
		FROM projects
		WHERE id = $1
	`

	var project Project
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, fmt.Errorf("get project by id: %w", err)
	}
	return project, nil
}

func (r *Repository) UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) (Project, error) {
	setClauses := make([]string, 0, 2)
	args := make([]any, 0, 3)
	nextArg := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", nextArg))
		args = append(args, strings.TrimSpace(*req.Name))
		nextArg++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", nextArg))
		args = append(args, strings.TrimSpace(*req.Description))
		nextArg++
	}

	args = append(args, projectID)
	query := fmt.Sprintf(`
		UPDATE projects
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, owner_id, created_at
	`, strings.Join(setClauses, ", "), nextArg)

	var updated Project
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.OwnerID,
		&updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return updated, nil
}

func (r *Repository) DeleteProject(ctx context.Context, projectID string) error {
	const query = `DELETE FROM projects WHERE id = $1`
	result, err := r.db.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
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

func (r *Repository) IsProjectOwner(ctx context.Context, projectID, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM projects WHERE id = $1 AND owner_id = $2)`
	var isOwner bool
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&isOwner); err != nil {
		return false, fmt.Errorf("check project owner: %w", err)
	}
	return isOwner, nil
}

func (r *Repository) ListProjectTasks(ctx context.Context, projectID string) ([]TaskSummary, error) {
	const query = `
		SELECT
			id,
			title,
			description,
			status::text,
			priority::text,
			project_id,
			assignee_id::text,
			created_by::text,
			COALESCE(to_char(due_date, 'YYYY-MM-DD'), ''),
			created_at,
			updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]TaskSummary, 0)
	for rows.Next() {
		task, err := scanTaskSummary(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project tasks rows: %w", err)
	}
	return tasks, nil
}

func (r *Repository) ListAvailableAssignees(ctx context.Context, _ string) ([]Assignee, error) {
	const query = `
		SELECT u.id::text, u.name, u.email
		FROM users u
		ORDER BY u.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list available assignees: %w", err)
	}
	defer rows.Close()

	assignees := make([]Assignee, 0)
	for rows.Next() {
		var a Assignee
		if err := rows.Scan(&a.ID, &a.Name, &a.Email); err != nil {
			return nil, fmt.Errorf("scan assignee: %w", err)
		}
		assignees = append(assignees, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assignee rows: %w", err)
	}

	return assignees, nil
}

func (r *Repository) GetProjectStats(ctx context.Context, projectID string) (StatsResponse, error) {
	statusRows, err := r.db.Query(ctx, `
		SELECT status::text, COUNT(*)
		FROM tasks
		WHERE project_id = $1
		GROUP BY status
	`, projectID)
	if err != nil {
		return StatsResponse{}, fmt.Errorf("query status stats: %w", err)
	}
	defer statusRows.Close()

	byStatus := make([]StatusCount, 0)
	for statusRows.Next() {
		var s StatusCount
		if err := statusRows.Scan(&s.Status, &s.Count); err != nil {
			return StatsResponse{}, fmt.Errorf("scan status stat: %w", err)
		}
		byStatus = append(byStatus, s)
	}
	if err := statusRows.Err(); err != nil {
		return StatsResponse{}, fmt.Errorf("iterate status stats: %w", err)
	}

	assigneeRows, err := r.db.Query(ctx, `
		SELECT
			COALESCE(u.id::text, 'unassigned'),
			COALESCE(u.name, 'Unassigned'),
			COUNT(*)
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		WHERE t.project_id = $1
		GROUP BY 1, 2
		ORDER BY 3 DESC, 2 ASC
	`, projectID)
	if err != nil {
		return StatsResponse{}, fmt.Errorf("query assignee stats: %w", err)
	}
	defer assigneeRows.Close()

	byAssignee := make([]AssigneeCount, 0)
	for assigneeRows.Next() {
		var a AssigneeCount
		if err := assigneeRows.Scan(&a.AssigneeID, &a.AssigneeName, &a.Count); err != nil {
			return StatsResponse{}, fmt.Errorf("scan assignee stat: %w", err)
		}
		byAssignee = append(byAssignee, a)
	}
	if err := assigneeRows.Err(); err != nil {
		return StatsResponse{}, fmt.Errorf("iterate assignee stats: %w", err)
	}

	return StatsResponse{ByStatus: byStatus, ByAssignee: byAssignee}, nil
}

func scanProject(row interface {
	Scan(dest ...any) error
}) (Project, error) {
	var project Project
	if err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	); err != nil {
		return Project{}, fmt.Errorf("scan project: %w", err)
	}
	return project, nil
}

func scanTaskSummary(row interface {
	Scan(dest ...any) error
}) (TaskSummary, error) {
	var task TaskSummary
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
		return TaskSummary{}, fmt.Errorf("scan task summary: %w", err)
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
