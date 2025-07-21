package server

import (
	"fmt"
	"net/http"
	"slices"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
}

// Pattern adopted from https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthy", healthy())
	mux.Handle("/", http.FileServer(http.Dir("static")))
	mux.Handle("/authenticate")

	var handler http.Handler = mux
	handler = authCheck(handler)
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

// Lets start with cookies https://www.calhoun.io/securing-cookies-in-go/
func authCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Auth hit")
		next.ServeHTTP(w, r)
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
