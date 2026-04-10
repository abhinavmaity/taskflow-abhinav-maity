package httpx

import "net/http"

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			WriteError(w, err)
		}
	}
}
