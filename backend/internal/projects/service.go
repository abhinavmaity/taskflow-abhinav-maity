package projects

import (
	"context"
	"errors"
	"strings"

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

func (s *Service) List(ctx context.Context, userID string, page, limit int) (ListProjectsResponse, error) {
	projects, total, err := s.repo.ListAccessibleProjects(ctx, userID, page, limit)
	if err != nil {
		return ListProjectsResponse{}, apperrors.WrapInternal(err)
	}

	return ListProjectsResponse{
		Projects:   projects,
		Pagination: httpx.NewPaginationMeta(page, limit, total),
	}, nil
}

func (s *Service) Create(ctx context.Context, userID string, req CreateProjectRequest) (Project, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return Project{}, apperrors.NewValidation(fields)
	}

	project := Project{
		ID:      uuid.NewString(),
		Name:    strings.TrimSpace(req.Name),
		OwnerID: userID,
	}
	if req.Description != nil {
		value := strings.TrimSpace(*req.Description)
		project.Description = &value
	}

	created, err := s.repo.CreateProject(ctx, project)
	if err != nil {
		return Project{}, apperrors.WrapInternal(err)
	}
	return created, nil
}

func (s *Service) Get(ctx context.Context, userID, projectID string) (ProjectDetailResponse, error) {
	project, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ProjectDetailResponse{}, apperrors.NewNotFound()
		}
		return ProjectDetailResponse{}, apperrors.WrapInternal(err)
	}

	allowed, err := s.repo.HasProjectAccess(ctx, projectID, userID)
	if err != nil {
		return ProjectDetailResponse{}, apperrors.WrapInternal(err)
	}
	if !allowed {
		return ProjectDetailResponse{}, apperrors.NewForbidden()
	}

	tasks, err := s.repo.ListProjectTasks(ctx, projectID)
	if err != nil {
		return ProjectDetailResponse{}, apperrors.WrapInternal(err)
	}

	assignees, err := s.repo.ListAvailableAssignees(ctx, projectID)
	if err != nil {
		return ProjectDetailResponse{}, apperrors.WrapInternal(err)
	}

	return ProjectDetailResponse{
		Project:            project,
		Tasks:              tasks,
		AvailableAssignees: assignees,
	}, nil
}

func (s *Service) Update(ctx context.Context, userID, projectID string, req UpdateProjectRequest) (Project, error) {
	if fields := req.Validate(); len(fields) > 0 {
		return Project{}, apperrors.NewValidation(fields)
	}

	existing, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Project{}, apperrors.NewNotFound()
		}
		return Project{}, apperrors.WrapInternal(err)
	}
	if existing.OwnerID != userID {
		return Project{}, apperrors.NewForbidden()
	}

	updated, err := s.repo.UpdateProject(ctx, projectID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Project{}, apperrors.NewNotFound()
		}
		return Project{}, apperrors.WrapInternal(err)
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, userID, projectID string) error {
	existing, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.NewNotFound()
		}
		return apperrors.WrapInternal(err)
	}
	if existing.OwnerID != userID {
		return apperrors.NewForbidden()
	}

	if err := s.repo.DeleteProject(ctx, projectID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apperrors.NewNotFound()
		}
		return apperrors.WrapInternal(err)
	}
	return nil
}

func (s *Service) Stats(ctx context.Context, userID, projectID string) (StatsResponse, error) {
	_, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return StatsResponse{}, apperrors.NewNotFound()
		}
		return StatsResponse{}, apperrors.WrapInternal(err)
	}

	allowed, err := s.repo.HasProjectAccess(ctx, projectID, userID)
	if err != nil {
		return StatsResponse{}, apperrors.WrapInternal(err)
	}
	if !allowed {
		return StatsResponse{}, apperrors.NewForbidden()
	}

	stats, err := s.repo.GetProjectStats(ctx, projectID)
	if err != nil {
		return StatsResponse{}, apperrors.WrapInternal(err)
	}
	return stats, nil
}
