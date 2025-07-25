package server

import (
	"net/http"
	"slices"

	"github.com/gorilla/sessions"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
}

type Config struct {
	MITID_CLIENT_ID     string
	MITID_CLIENT_SECRET string
	ENVIRONMENT         string
	TOKEN_SIGN_KEY      string
	SESSION_KEY         string
}

// Pattern adopted from https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer(config Config) http.Handler {
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))
	var handler http.Handler = addRoutes(store, config)
	// we want cors check first, because that is the simplest access check
	handler = checkCORS(handler)

	return handler
}

// refactored into independent route function to aid testing
func addRoutes(store *sessions.CookieStore, config Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/healthy", healthy())

	// Handles full authentication
	mux.Handle("/account/authenticate", authenticate(store, config))
	mux.Handle("/account/login", login(config))
	mux.Handle("/account/onboarding", onboarding(store, config))

	/* Then we need something that handles all incoming */

	return mux
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
