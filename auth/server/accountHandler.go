package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/henrikkorsgaard/gaia/auth/tokens"
	"github.com/henrikkorsgaard/gaia/crm/database"
)

type mitidTokens struct {
	IdToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
}

type mitidUser struct {
	MitIdUUID string `json:"mitid.uuid"`
	Name      string `json:"mitid.identity_name"`
}

var state = uuid.NewString()[:6]

var (
	ErrAuthenticationMissingCode         = errors.New("error: mitid did not return code.")
	ErrAuthenticationStateError          = errors.New("error: provider returned unexpected state.")
	ErrAuthenticationIdentityNotFound    = errors.New("error: crm could not match identity.")
	ErrAuthenticationIdentityServiceFail = errors.New("error: crm returned error.")
	ErrAuthenticationOnboardingSession   = errors.New("error: onbaording sessions data incomplete.")
)

/*
login() will redirect the user to MitID authentication flow
this will redirect here with the codes needed.
*/
func authenticate(store *sessions.CookieStore) http.Handler {
	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		queryState := q.Get("state")
		if queryState != state {
			http.Error(w, ErrAuthenticationStateError.Error(), http.StatusInternalServerError)
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Error(w, ErrAuthenticationMissingCode.Error(), http.StatusInternalServerError)
			return
		}

		// This can potentially fail if the code is old?
		mtokens, err := getTokens(code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//This can fail if the login was initiated more than 15 minutes earlier
		mitUser, err := getUserInfo(mtokens.AccessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := matchUser(mitUser.MitIdUUID, mitUser.Name, "", "")
		//We handle errors here that are not associated with 404 identity match
		//This returns any other error
		if err != nil && !errors.Is(err, ErrAuthenticationIdentityNotFound) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user.GaiaId != "" {
			token, err := tokens.NewUserToken(user.GaiaId)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

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
			//TODO: Handle redirect targets in config
			http.Redirect(w, r, "/gaia/dashboard.html", http.StatusFound)
		} else {
			session, err := store.Get(r, "gaia")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			session.Values["mitid"] = mitUser

			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			cookie := http.Cookie{
				Name:     "gaia_n",
				Value:    mitUser.Name,
				MaxAge:   3600,
				Path:     "/",
				HttpOnly: false,
				Secure:   false,
			}

			http.SetCookie(w, &cookie)

			//TODO: Handle this from a config perspective.
			http.Redirect(w, r, "/onboarding.html", http.StatusTemporaryRedirect)
		}
	})
}

func onboarding(store *sessions.CookieStore) http.Handler {

	//this is the endpoint that sets what?
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {
			session, err := store.Get(r, "gaia")
			if err != nil {
				clearOnboardingSessionData(w, r, session)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			user, err := getOnboardingSessionUser(r, session)
			if err != nil {
				clearOnboardingSessionData(w, r, session)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			user, err = matchUser(user.MitIdUUID, user.Name, user.Address, user.DarId)
			if err != nil {
				clearOnboardingSessionData(w, r, session)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if user.GaiaId == "" {
				clearOnboardingSessionData(w, r, session)
				http.Error(w, ErrAuthenticationIdentityNotFound.Error(), http.StatusInternalServerError)
				return
			}

			token, err := tokens.NewUserToken(user.GaiaId)
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

			http.Redirect(w, r, "/gaia/dashboard.html", http.StatusOK)
		}

	})
}

func login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//TODO: manage host in config
		uri := "http://localhost:3020/authenticate"
		clientId := os.Getenv("MITID_CLIENT_ID")
		simulated := "" // we use this to auto sign in when doing development.
		//TODO handle dev environment better
		if os.Getenv("ENVIRONMENT") == "dev" {
			simulated = "&simulation=no-ui uuid:0e4a1734-a8f3-4c49-b09c-35405104725e"
		}

		url := fmt.Sprintf("https://pp.netseidbroker.dk/op/connect/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid mitid&state=%s%s", clientId, uri, state, simulated)
		http.Redirect(w, r, url, http.StatusFound)
	})
}

/*
This is the proxy middleware for authentication checks.
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

func getTokens(code string) (mtokens mitidTokens, err error) {

	// Now we go on to exchaning the code for access and id tokens
	data := url.Values{}
	data.Add("client_id", os.Getenv("MITID_CLIENT_ID"))
	data.Add("client_secret", os.Getenv("MITID_CLIENT_SECRET"))
	data.Add("grant_type", "authorization_code")
	data.Add("code", code)
	data.Add("redirect_uri", "http://localhost:3000/authenticate")

	resp, err := http.Post("https://pp.netseidbroker.dk/op/connect/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return mtokens, err
	}
	// we need to close this, see: https://stackoverflow.com/questions/23928983/defer-body-close-after-receiving-response
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&mtokens)

	return mtokens, err
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

func matchUser(mitidUUID, name, address, darId string) (user database.User, err error) {
	var data = fmt.Sprintf(`{ "mitid_uuid":"%s", "name":"%s", "address": "%s", "dar_id": "%s" }`, mitidUUID, name, address, darId)
	//TODO: Manage CRM host in config
	resp, err := http.Post("http://localhost:3010/match", "application/json", strings.NewReader(data))
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {

		return user, ErrAuthenticationIdentityNotFound
	} else if resp.StatusCode == http.StatusOK {
		json.NewDecoder(resp.Body).Decode(&user)
		return user, err
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return user, errors.Join(ErrAuthenticationIdentityServiceFail, err)
		}

		return user, errors.Join(ErrAuthenticationIdentityServiceFail, errors.New(string(body)))
	}
}

func getOnboardingSessionUser(r *http.Request, session *sessions.Session) (user database.User, err error) {
	address := r.FormValue("address")
	darId := r.FormValue("darid")

	if address == "" || darId == "" {
		return user, ErrAuthenticationOnboardingSession
	}

	mitid := session.Values["mitid"].(mitidUser)

	if session.IsNew || mitid.MitIdUUID == "" || mitid.Name == "" {
		return user, ErrAuthenticationOnboardingSession
	}

	cookie, err := r.Cookie("gaia_n")
	if err != nil {
		return user, err
	}

	//If we cannot match the values between name and id,
	// then something is tampered with
	if cookie.Value == "" || cookie.Value != mitid.Name {
		return user, ErrAuthenticationOnboardingSession
	}

	user = database.User{
		MitIdUUID: mitid.MitIdUUID,
		Name:      mitid.Name,
		Address:   address,
		DarId:     darId,
	}

	return user, err
}

func clearOnboardingSessionData(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	//remove data from session
	delete(session.Values, "mitid")

	cookie := http.Cookie{
		Name:     "gaia_n",
		Value:    "",
		MaxAge:   0,
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
	}

	http.SetCookie(w, &cookie)
	return session.Save(r, w)
}
