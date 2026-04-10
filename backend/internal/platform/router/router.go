package router

import (
	"log/slog"
	"net/http"

	"github.com/abhinavmaity/taskflow/backend/internal/auth"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/authctx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/config"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/middleware"
	"github.com/abhinavmaity/taskflow/backend/internal/projects"
	"github.com/abhinavmaity/taskflow/backend/internal/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(logger *slog.Logger, cfg config.Config, dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CORS(cfg.CORSOrigin))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.RequestLogger(logger))

	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
	authRepo := auth.NewRepository(dbPool)
	authService := auth.NewService(authRepo, tokenManager)
	authHandler := auth.NewHandler(authService)
	projectsRepo := projects.NewRepository(dbPool)
	projectsService := projects.NewService(projectsRepo)
	projectsHandler := projects.NewHandler(projectsService)
	tasksRepo := tasks.NewRepository(dbPool)
	tasksService := tasks.NewService(tasksRepo)
	tasksHandler := tasks.NewHandler(tasksService)

	r.Get("/health", httpx.Handle(func(w http.ResponseWriter, _ *http.Request) error {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return nil
	}))
	authHandler.RegisterRoutes(r)

	r.Group(func(protected chi.Router) {
		protected.Use(middleware.RequireAuth(tokenManager))

		protected.Get("/me", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
			user, ok := authctx.CurrentUserFromContext(r.Context())
			if !ok {
				return apperrors.NewUnauthorized()
			}
			httpx.WriteJSON(w, http.StatusOK, map[string]any{"user": user})
			return nil
		}))

		projectsHandler.RegisterRoutes(protected)
		tasksHandler.RegisterRoutes(protected)
	})

	r.NotFound(httpx.Handle(func(_ http.ResponseWriter, _ *http.Request) error {
		return apperrors.NewNotFound()
	}))

	return r
}
