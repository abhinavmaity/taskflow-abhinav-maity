package router

import (
	"log/slog"
	"net/http"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/authctx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/config"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(logger *slog.Logger, cfg config.Config, _ *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.RequestLogger(logger))

	r.Get("/health", httpx.Handle(func(w http.ResponseWriter, _ *http.Request) error {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return nil
	}))

	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.RequireAuth(cfg.JWTSecret))

		api.Get("/me", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			user, ok := authctx.CurrentUserFromContext(r.Context())
			if !ok {
				return apperrors.NewUnauthorized()
			}
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
			return nil
		}))

		// Temporary endpoint for verifying 403 mapping before domain modules exist.
		api.Get("/forbidden-check", httpx.Handle(func(_ http.ResponseWriter, _ *http.Request) error {
			return apperrors.NewForbidden()
		}))
	})

	r.NotFound(httpx.Handle(func(_ http.ResponseWriter, _ *http.Request) error {
		return apperrors.NewNotFound()
	}))

	return r
}
