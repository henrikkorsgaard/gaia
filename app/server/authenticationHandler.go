package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type tokens struct {
	IdToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
}

// https://github.com/gorilla/securecookie
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

type mitidUser struct {
	MitIdUUID string `json:"mitid.uuid"`
	Name      string `json:"mitid.identity_name"`
}

var state = uuid.NewString()[:6]

var (
	ErrAuthenticationMissingCode = "error: mitid did not return code."
	ErrAuthenticationStateError  = "error: provider returned unexpected state."
)

/*
login() will redirect the user to MitID authentication flow
this will redirect here with the codes needed.
*/
func authenticate() http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		queryState := q.Get("state")
		if queryState != state {
			http.Error(w, ErrAuthenticationStateError, http.StatusInternalServerError)
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Error(w, ErrAuthenticationMissingCode, http.StatusInternalServerError)
			return
		}

		// This can potentially fail if the code is old?
		tokens, err := getTokens(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//This can fail if the login was initiated more than 15 minutes earlier
		mitUser, err := getUserInfo(tokens.AccessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := matchUser(mitUser.MitIdUUID, mitUser.Name, "", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if token != "" {
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["token"] = token
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			//http cookie for name
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["mit"] = mitUser.MitIdUUID
			session.Values["name"] = mitUser.Name

			err = session.Save(r, w)

			cookie := http.Cookie{
				Name:     "gaia_n",
				Value:    mitUser.Name,
				MaxAge:   3600,
				Path:     "/",
				HttpOnly: false,
				Secure:   false,
			}

			http.SetCookie(w, &cookie)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/onboarding.html", http.StatusSeeOther)
		}
	})
}

func onboarding() http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//If cookie not present -> error

		//If form data is not completed -> error

		//Get token

		//if succesful -> set cookie and redirect

		//if unsucesful -> redirect to error page

		token, err := matchUser(mitUser.MitIdUUID, mitUser.Name, "", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if token != "" {
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["token"] = token
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			//http cookie for name
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["mit"] = mitUser.MitIdUUID
			session.Values["name"] = mitUser.Name

			err = session.Save(r, w)

			cookie := http.Cookie{
				Name:     "gaia_n",
				Value:    mitUser.Name,
				MaxAge:   3600,
				Path:     "/",
				HttpOnly: false,
				Secure:   false,
			}

			http.SetCookie(w, &cookie)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/onboarding.html", http.StatusSeeOther)
		}
	})
}

func login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		uri := "http://localhost:3000/authenticate"
		clientId := os.Getenv("MITID_CLIENT_ID")
		simulated := "" // we use this to auto sign in when doing development.
		//TODO handle dev environment better
		if os.Getenv("ENVIRONMENT") == "dev" {
			simulated = "&simulation=no-ui uuid:0e4a1734-a8f3-4c49-b09c-35405104725e"
		}

		url := fmt.Sprintf("https://pp.netseidbroker.dk/op/connect/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid mitid&state=%s%s", clientId, uri, state, simulated)
		http.Redirect(w, r, url, http.StatusSeeOther)
	})
}

/*
This middleware should be used to check authentication for protected endpoints
The methodology (secure cookie vs JWT token is still not determined)
*/
func authCheck(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth checker should check for token
		// if no token redirect to login
		//
		/*
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			/*
			if session.IsNew && !strings.Contains(r.URL.Path, "/authenticate") {
				http.Redirect(w, r, "/authenticate", http.StatusSeeOther)
				return
			}*/

		next.ServeHTTP(w, r)
	})
}

func getTokens(code string) (tokens tokens, err error) {

	// Now we go on to exchaning the code for access and id tokens
	data := url.Values{}
	data.Add("client_id", os.Getenv("MITID_CLIENT_ID"))
	data.Add("client_secret", os.Getenv("MITID_CLIENT_SECRET"))
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", "http://localhost:3000/authenticate")

	resp, err := http.Post("https://pp.netseidbroker.dk/op/connect/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return tokens, err
	}
	// we need to close this, see: https://stackoverflow.com/questions/23928983/defer-body-close-after-receiving-response
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&tokens)

	return tokens, err
}

func getUserInfo(accessToken string) (user mitidUser, err error) {

	req, err := http.NewRequest("GET", "https://pp.netseidbroker.dk/op/connect/userinfo", nil)
	if err != nil {
		return user, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&user)

	return user, err
}

func matchUser(mitidUUID, name, address, darId string) (token string, err error) {
	var data = fmt.Sprintf(`{ "mitid_uuid":"%s", "name":"%s", "address": "%s", "dar_id": "%s" }`, mitidUUID, name, address, darId)
	resp, err := http.Post("http://localhost:3010/match", "application/json", strings.NewReader(data))
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return token, err
	}

	if resp.StatusCode == http.StatusOK {
		token = string(body)
	}

	return token, err
}
