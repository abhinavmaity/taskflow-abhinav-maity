package tasks

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID, projectID string, filters TaskFilters, page, limit int) (TaskListResponse, error) {
	projectExists, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return TaskListResponse{}, apperrors.WrapInternal(err)
	}
	if !projectExists {
		return TaskListResponse{}, apperrors.NewNotFound()
	}

	allowed, err := s.repo.HasProjectAccess(ctx, projectID, userID)
	if err != nil {
		return TaskListResponse{}, apperrors.WrapInternal(err)
	}
	if !allowed {
		return TaskListResponse{}, apperrors.NewForbidden()
	}

	tasks, total, err := s.repo.ListTasks(ctx, projectID, filters, page, limit)
	if err != nil {
		return TaskListResponse{}, apperrors.WrapInternal(err)
	}

	return TaskListResponse{
		Tasks:      tasks,
		Pagination: httpx.NewPaginationMeta(page, limit, total),
	}, nil
}

func (s *Service) Create(ctx context.Context, userID, projectID string, req CreateTaskRequest) (Task, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return Task{}, apperrors.NewValidation(fields)
	}

	if err := validateOptionalDueDate(req.DueDate); err != nil {
		return Task{}, err
	}

	projectExists, err := s.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return Task{}, apperrors.WrapInternal(err)
	}
	if !projectExists {
		return Task{}, apperrors.NewNotFound()
	}

	allowed, err := s.repo.HasProjectAccess(ctx, projectID, userID)
	if err != nil {
		return Task{}, apperrors.WrapInternal(err)
	}
	if !allowed {
		return Task{}, apperrors.NewForbidden()
	}

	assigneeID, err := s.validateAssignee(ctx, req.AssigneeID)
	if err != nil {
		return Task{}, err
	}

	status := req.Status
	if status == "" {
		status = "todo"
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	task := Task{
		ID:          uuid.NewString(),
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		ProjectID:   projectID,
		AssigneeID:  assigneeID,
		CreatedBy:   userID,
		DueDate:     req.DueDate,
	}

	created, err := s.repo.CreateTask(ctx, task)
	if err != nil {
		return Task{}, apperrors.WrapInternal(err)
	}
	return created, nil
}

func (s *Service) Update(ctx context.Context, userID, taskID string, req UpdateTaskRequest) (Task, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return Task{}, apperrors.NewValidation(fields)
	}

	if err := validateOptionalDueDate(req.DueDate); err != nil {
		return Task{}, err
	}

	perm, err := s.repo.GetTaskPermissionContext(ctx, taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Task{}, apperrors.NewNotFound()
		}
		return Task{}, apperrors.WrapInternal(err)
	}

	allowed, err := s.repo.HasProjectAccess(ctx, perm.ProjectID, userID)
	if err != nil {
		return Task{}, apperrors.WrapInternal(err)
	}
	if !allowed {
		return Task{}, apperrors.NewForbidden()
	}

	if _, err := s.validateAssignee(ctx, req.AssigneeID); err != nil {
		return Task{}, err
	}

	updated, err := s.repo.UpdateTask(ctx, taskID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Task{}, apperrors.NewNotFound()
		}
		return Task{}, apperrors.WrapInternal(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, userID, taskID string) error {
	perm, err := s.repo.GetTaskPermissionContext(ctx, taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.NewNotFound()
		}
		return apperrors.WrapInternal(err)
	}

	if userID != perm.ProjectOwner && userID != perm.CreatedBy {
		return apperrors.NewForbidden()
	}

	if err := s.repo.DeleteTask(ctx, taskID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.NewNotFound()
		}
		return apperrors.WrapInternal(err)
	}
	return nil
}

func (s *Service) validateAssignee(ctx context.Context, assigneeID *string) (*string, error) {
	if assigneeID == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*assigneeID)
	if trimmed == "" {
		return nil, nil
	}

	exists, err := s.repo.UserExists(ctx, trimmed)
	if err != nil {
		return nil, apperrors.WrapInternal(err)
	}
	if !exists {
		return nil, apperrors.NewValidation(map[string]string{
			"assignee_id": "must reference an existing user",
		})
	}
	return &trimmed, nil
}

func validateOptionalDueDate(dueDate *string) error {
	if dueDate == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*dueDate)
	if trimmed == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return apperrors.NewValidation(map[string]string{
			"due_date": "must be in YYYY-MM-DD format",
		})
	}
	return nil
}
