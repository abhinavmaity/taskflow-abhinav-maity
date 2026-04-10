package tasks

import (
	"net/http"
	"strings"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/authctx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/projects/{id}/tasks", httpx.Handle(h.listByProject))
	r.Post("/projects/{id}/tasks", httpx.Handle(h.create))
	r.Patch("/tasks/{id}", httpx.Handle(h.update))
	r.Delete("/tasks/{id}", httpx.Handle(h.delete))
}

func (h *Handler) listByProject(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	pagination, err := httpx.ParsePagination(r)
	if err != nil {
		return err
	}

	filters := TaskFilters{
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
		Assignee: strings.TrimSpace(r.URL.Query().Get("assignee")),
	}
	if filters.Status != "" {
		if _, ok := validStatuses[filters.Status]; !ok {
			return apperrors.NewValidation(map[string]string{
				"status": "must be one of: todo, in_progress, done",
			})
		}
	}

	resp, err := h.service.List(r.Context(), user.ID, projectID, filters, pagination.Page, pagination.Limit)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	var req CreateTaskRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	created, err := h.service.Create(r.Context(), user.ID, projectID, req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusCreated, created)
	return nil
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	taskID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	var req UpdateTaskRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	updated, err := h.service.Update(r.Context(), user.ID, taskID, req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, updated)
	return nil
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	taskID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), user.ID, taskID); err != nil {
		return err
	}

	httpx.WriteNoContent(w)
	return nil
}
