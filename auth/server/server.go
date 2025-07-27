package server

import (
	"log"
	"net/http"
	"net/url"
	"slices"

	"github.com/gorilla/sessions"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
}

type Config struct {
	ENVIRONMENT string `env:"ENVIRONMENT,required"`
	//Mitid Broker
	MITID_CLIENT_ID     string `env:"MITID_CLIENT_ID,required"`
	MITID_CLIENT_SECRET string `env:"MITID_CLIENT_SECRET,required"`
	MITID_BROKER_HOST   string `env:"MITID_BROKER_HOST,required"`
	//Keys
	TOKEN_SIGN_KEY string `env:"TOKEN_SIGN_KEY,required"`
	SESSION_KEY    string `env:"SESSION_KEY,required"`
	//Hosts
	ORIGIN_SERVER string `env:"ORIGIN_SERVER,required"`
	CRM_SERVER    string `env:"CRM_SERVER,required"`
	//Redirects
	POST_LOGIN_REDIRECT        string `env:"POST_LOGIN_REDIRECT"`
	IDENTITY_ERROR_REDIRECT    string `env:"IDENTITY_ERROR_REDIRECT"`
	AUTH_SERVER_ERROR_REDIRECT string `env:"AUTH_SERVER_ERROR_REDIRECT"`
}

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

	originServer, err := url.Parse(config.ORIGIN_SERVER)

	if err != nil {
		log.Fatal("invalid origin server URL")
	}
	mux.Handle("/", proxyHandler(store, originServer, config))

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
