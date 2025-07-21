package server

import (
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/gorilla/sessions"
	_ "github.com/joho/godotenv/autoload"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
}

// https://github.com/gorilla/securecookie

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

// Pattern adopted from https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer() http.Handler {

	mux := http.NewServeMux()
	mux.Handle("/healthy", healthy())
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

// This is the endpoint for external authentication redirect url
func authenticate() http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get a session. Get() always returns a session, even if empty.
		session, err := store.Get(r, "gaia")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set some session values.
		session.Values["foo"] = "bar"
		session.Values[42] = 43
		// Save it before we write to the response/return from the handler.
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

// Consider authenticating on all endpoints that needs authentication
// Authorizing should happen on each API request
func authCheck(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "gaia")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if session.IsNew && !strings.Contains(r.URL.Path, "/authenticate") {
			http.Redirect(w, r, "/authenticate", http.StatusSeeOther)
			return
		}

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
