package server

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
	"http://localhost:80",
	"http://127.0.0.1:80",
	"http://127.0.0.1:3010",
	"http://localhost:3010",
}

// TODO: Add filepath for static folder to make this testable
func NewServer() http.Handler {

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("../static")))
	mux.Handle("/healthy", healthy())

	var handler http.Handler = mux
	handler = ignoreFavicon(handler)
	// we want cors check first, because that is the simplest access check
	// likely moved to the auth gateway
	handler = checkCORS(handler)
	handler = staticLog(handler)

	return handler
}

func healthy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte("I'm alive"))
	})
}

func staticLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[Static server] received for path %s request at: %s\n", r.URL.Path, time.Now())
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

func ignoreFavicon(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "favicon.ico") {
			http.NotFoundHandler()
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
