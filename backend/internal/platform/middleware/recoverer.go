package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
)

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"request_id", RequestIDFromContext(r.Context()),
						"panic", fmt.Sprintf("%v", rec),
					)
					httpx.WriteJSON(w, http.StatusInternalServerError, httpx.ErrorResponse{
						Error: "internal server error",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
