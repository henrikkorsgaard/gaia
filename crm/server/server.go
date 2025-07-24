package server

import (
	"net/http"
	"slices"

	"github.com/henrikkorsgaard/gaia/crm/database"
)

// Should be put in .env
var originAllowlist = []string{
	"http://127.0.0.1:8000",
	"http://localhost:8000",
}

// Pattern adopted from https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func NewServer(db *database.UserDatabase) http.Handler {

	var handler http.Handler = addRoutes(db)
	// we want cors check first, because that is the simplest access check
	handler = checkCORS(handler)

	return handler
}

// refactored into independent route function to aid testing
func addRoutes(db *database.UserDatabase) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/healthy", healthy())
	//Returns JSON
	mux.Handle("/users/{id}", userIdHandler(db)) //GET, PUT, POST, DELETE
	mux.Handle("/users", userHandler(db))        //GET List
	mux.Handle("/match", matchHandler(db))
	mux.Handle("/", viewHandler(db))
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
