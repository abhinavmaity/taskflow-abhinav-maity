package auth

import (
	"net/http"

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
	r.Post("/auth/register", httpx.Handle(h.register))
	r.Post("/auth/login", httpx.Handle(h.login))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
	return nil
}
