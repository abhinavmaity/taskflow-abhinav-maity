package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/abhinavmaity/taskflow/backend/internal/platform/apperrors"
)

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteError(w http.ResponseWriter, err error) {
	status, body := MapError(err)
	WriteJSON(w, status, body)
}

func MapError(err error) (int, ErrorResponse) {
	appErr, ok := apperrors.As(err)
	if !ok {
		return http.StatusInternalServerError, ErrorResponse{Error: "internal server error"}
	}

	switch appErr.Kind {
	case apperrors.KindValidation:
		return http.StatusBadRequest, ErrorResponse{
			Error:  "validation failed",
			Fields: appErr.Fields,
		}
	case apperrors.KindUnauthorized:
		return http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"}
	case apperrors.KindForbidden:
		return http.StatusForbidden, ErrorResponse{Error: "forbidden"}
	case apperrors.KindNotFound:
		return http.StatusNotFound, ErrorResponse{Error: "not found"}
	default:
		return http.StatusInternalServerError, ErrorResponse{Error: "internal server error"}
	}
}
