package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// This is the endpoint for external authentication redirect url
func authenticate() http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		code := q.Get("code")
		fmt.Println(code)
		data := url.Values{}
		data.Add("client_id", os.Getenv("MITID_CLIENT_ID"))
		data.Add("client_secret", os.Getenv("MITID_CLIENT_SECRET"))
		data.Add("grant_type", "authorization_code")
		data.Add("code", code)
		data.Add("redirect_uri", "http://localhost:3000/authenticate")

		resp, err := http.Post("https://pp.netseidbroker.dk/op/connect/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error while reading the response bytes:", err)
		}
		fmt.Println("H " + string([]byte(body)))

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}

func login() http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// we get info from MitId
		// can we decode it here (or should we move it elsewhere?)
		//

		// Get a session. Get() always returns a session, even if empty.
		// Should we check the header?
		/*'
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
		*/
		// This allow us to manipulate state
		// Test identity: Victoria43276, uuid:0e4a1734-a8f3-4c49-b09c-35405104725e
		url := "https://pp.netseidbroker.dk/op/connect/authorize?response_type=code&client_id=" + os.Getenv("MITID_CLIENT_ID") + "&redirect_uri=http://localhost:3000/authenticate&scope=openid mitid&state=xyz123&simulation=no-ui uuid:0e4a1734-a8f3-4c49-b09c-35405104725e"
		http.Redirect(w, r, url, http.StatusSeeOther)
	})
}

// Consider authenticating on all endpoints that needs authentication
// Authorizing should happen on each API request
func authCheck(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth checker should check for token
		// if no token redirect to login
		//
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
