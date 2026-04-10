package projects

import (
	"net/http"

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
	r.Get("/projects", httpx.Handle(h.list))
	r.Post("/projects", httpx.Handle(h.create))
	r.Get("/projects/{id}", httpx.Handle(h.get))
	r.Patch("/projects/{id}", httpx.Handle(h.update))
	r.Delete("/projects/{id}", httpx.Handle(h.delete))
	r.Get("/projects/{id}/stats", httpx.Handle(h.stats))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	pagination, err := httpx.ParsePagination(r)
	if err != nil {
		return err
	}

	resp, err := h.service.List(r.Context(), user.ID, pagination.Page, pagination.Limit)
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

	var req CreateProjectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	created, err := h.service.Create(r.Context(), user.ID, req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusCreated, created)
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	resp, err := h.service.Get(r.Context(), user.ID, projectID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	var req UpdateProjectRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	updated, err := h.service.Update(r.Context(), user.ID, projectID, req)
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

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), user.ID, projectID); err != nil {
		return err
	}

	httpx.WriteNoContent(w)
	return nil
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) error {
	user, ok := authctx.CurrentUserFromContext(r.Context())
	if !ok {
		return apperrors.NewUnauthorized()
	}

	projectID, err := httpx.ParseUUIDParam(r, "id")
	if err != nil {
		return err
	}

	resp, err := h.service.Stats(r.Context(), user.ID, projectID)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
	return nil
}
