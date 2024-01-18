package handlers

import (
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

func trimEnd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// очистка от конечных пробелов
		r.URL.Path = strings.TrimSpace(r.URL.Path)
		// очистка от конечных слешей
		r.URL.Path = strings.TrimRight(r.URL.Path, "/")

		next.ServeHTTP(w, r)
	})
}
