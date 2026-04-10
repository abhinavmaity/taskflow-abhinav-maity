package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/abhinavmaity/taskflow/backend/internal/auth"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/authctx"
	"github.com/abhinavmaity/taskflow/backend/internal/platform/httpx"
)

func RequireAuth(tokenManager *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				httpx.WriteError(w, apperrors.NewUnauthorized())
				return
			}

			claims, err := tokenManager.Parse(tokenString)
			if err != nil {
				httpx.WriteError(w, apperrors.NewUnauthorized())
				return
			}

			user := authctx.CurrentUser{
				ID:    claims.UserID,
				Email: claims.Email,
			}

			next.ServeHTTP(w, r.WithContext(authctx.WithCurrentUser(r.Context(), user)))
		})
	}
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid authorization header")
	}

	return strings.TrimSpace(parts[1]), nil
}
