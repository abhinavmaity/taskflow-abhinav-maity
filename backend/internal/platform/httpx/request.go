package httpx

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaginationQuery struct {
	Page  int
	Limit int
}

func DecodeJSON(r *http.Request, out any) error {
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

func ParsePagination(r *http.Request) (PaginationQuery, error) {
	page, err := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		return PaginationQuery{}, apperrors.NewValidation(map[string]string{
			"page": "must be a positive integer",
		})
	}

	limit, err := parsePositiveInt(r.URL.Query().Get("limit"), 10)
	if err != nil {
		return PaginationQuery{}, apperrors.NewValidation(map[string]string{
			"limit": "must be a positive integer",
		})
	}

	if limit > 100 {
		return PaginationQuery{}, apperrors.NewValidation(map[string]string{
			"limit": "must be less than or equal to 100",
		})
	}

	return PaginationQuery{Page: page, Limit: limit}, nil
}

func ParseUUIDParam(r *http.Request, name string) (string, error) {
	value := chi.URLParam(r, name)
	if value == "" {
		return "", apperrors.NewValidation(map[string]string{name: "is required"})
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", apperrors.NewValidation(map[string]string{name: "must be a valid UUID"})
	}
	return value, nil
}

func parsePositiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, err
	}
	return value, nil
}
