package server

import (
	"net/http"
	"slices"
)

// Pattern adopted from https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer() http.Handler {

	mux := http.NewServeMux()
	mux.Handle("/healthy", healthy())
	//mux.Handle("/user/{id}")

	var handler http.Handler = mux
	// we want cors check first, because that is the simplest access check
	handler = checkCORS(handler)

	return handler
}

func healthy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("I'm alive"))
	})
}

func checkCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if slices.Contains(originAllowlist, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)

		}
		w.Header().Add("Vary", "Origin")
		next.ServeHTTP(w, r)
	})
}
