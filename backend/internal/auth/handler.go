package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
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
	if err := decodeJSON(r, &req); err != nil {
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
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
	return nil
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return apperrors.NewValidation(map[string]string{
			"body": "must be valid JSON",
		})
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return apperrors.NewValidation(map[string]string{
				"body": "must contain a single JSON object",
			})
		}
		return apperrors.NewValidation(map[string]string{
			"body": "must be valid JSON",
		})
	}

	return nil
}
