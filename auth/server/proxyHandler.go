package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/henrikkorsgaard/gaia/auth/tokens"
)

var (
	ErrInvalidSession = errors.New("invalid authentication session")
	ErrInvalidToken   = errors.New("invalid authentication token")
)

func proxyHandler(store *sessions.CookieStore, originServer *url.URL, config Config) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//TODO: Maybe generalise some pattern matching here?
		//TODO: Integrate with a configuration setting, e.g. map of paths
		if strings.HasPrefix(r.URL.Path, "/secret/") {

			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tokenString := session.Values["token"]
			if session.IsNew || tokenString == nil || tokenString.(string) == "" {
				http.Error(w, ErrInvalidSession.Error(), http.StatusUnauthorized)
				return
			}

			validToken, err := tokens.ValidateToken(session.Values["token"].(string), config.TOKEN_SIGN_KEY)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !validToken {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}
		}

		r.Host = originServer.Host
		r.URL.Host = originServer.Host
		r.URL.Scheme = originServer.Scheme
		r.RequestURI = ""

		// save the response from the origin server
		oResp, err := http.DefaultClient.Do(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}
		defer oResp.Body.Close()

		// return response to the client
		w.WriteHeader(http.StatusOK)
		io.Copy(w, oResp.Body)
	})
}
