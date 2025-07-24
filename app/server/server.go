package server

import (
	"net/http"
	"slices"
	"strings"

	_ "github.com/joho/godotenv/autoload"
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
	mux.Handle("/login", login())
	mux.Handle("/authenticate", authenticate())
	mux.Handle("/", http.FileServer(http.Dir("static")))

	var handler http.Handler = mux
	handler = ignoreFavicon(handler)
	// we want cors check first, because that is the simplest access check
	handler = checkCORS(handler)
	// authCheck will check the cookie and then redirect to /authenticate if no cookie is found
	handler = authCheck(handler)

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

func ignoreFavicon(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "favicon.ico") {
			http.NotFoundHandler()
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
