package middleware

import (
	"net/http"
	"strings"
)

const (
	allowHeaders = "Authorization, Content-Type"
	allowMethods = "GET, POST, PATCH, DELETE, OPTIONS"
)

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	allowedOrigin = strings.TrimSpace(allowedOrigin)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" && isAllowedOrigin(origin, allowedOrigin) {
				if allowedOrigin == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin, allowedOrigin string) bool {
	if allowedOrigin == "" {
		return false
	}
	if allowedOrigin == "*" {
		return true
	}
	return strings.EqualFold(origin, allowedOrigin)
}
