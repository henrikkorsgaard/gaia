package server

import (
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/henrikkorsgaard/gaia/auth/tokens"
	"github.com/henrikkorsgaard/gaia/crm/database"
	"github.com/henrikkorsgaard/gaia/crm/server"
	"github.com/matryer/is"
)

var testdb = "test.db"

func TestCookieAuthCheck(t *testing.T) {
	is := is.New(t)

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello world")
	}))

	defer source.Close()
	u := database.User{
		GaiaId: uuid.New().String(),
	}

	config := getServerConfig()
	config.FRONTEND_SERVER = source.URL
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))

	authServer := httptest.NewServer(addRoutes(store, config))
	defer authServer.Close()

	token, err := tokens.NewUserToken(u.GaiaId, config.TOKEN_SIGN_KEY)
	is.NoErr(err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/secret/page.html", authServer.URL), nil)
	is.NoErr(err)

	session, err := store.Get(req, "gaia")
	is.NoErr(err)

	session.Values["token"] = token

	recorder := httptest.NewRecorder()
	err = session.Save(req, recorder)
	is.NoErr(err)

	req.AddCookie(recorder.Result().Cookies()[0])

	client := authServer.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, http.StatusOK)
}

func TestIndexWithCookie(t *testing.T) {
	is := is.New(t)

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello world")
	}))
	defer source.Close()

	u := database.User{
		GaiaId: uuid.New().String(),
	}

	config := getServerConfig()
	config.FRONTEND_SERVER = source.URL
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))

	authServer := httptest.NewServer(addRoutes(store, config))
	defer authServer.Close()

	token, err := tokens.NewUserToken(u.GaiaId, config.TOKEN_SIGN_KEY)
	is.NoErr(err)

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/index.html", authServer.URL), nil)
	is.NoErr(err)

	session, err := store.Get(req, "gaia")
	is.NoErr(err)

	session.Values["token"] = token

	recorder := httptest.NewRecorder()
	err = session.Save(req, recorder)
	is.NoErr(err)

	req.AddCookie(recorder.Result().Cookies()[0])

	client := authServer.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, http.StatusOK)
}

func TesMissingCookieFail(t *testing.T) {
	is := is.New(t)

	config := getServerConfig()
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))

	authServer := httptest.NewServer(addRoutes(store, config))
	defer authServer.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/secret/page.html", authServer.URL), nil)
	is.NoErr(err)

	client := authServer.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, http.StatusUnauthorized)
}

func TestMissingTokenFail(t *testing.T) {
	is := is.New(t)

	config := getServerConfig()
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))

	authServer := httptest.NewServer(addRoutes(store, config))
	defer authServer.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/secret/page.html", authServer.URL), nil)
	is.NoErr(err)

	session, err := store.Get(req, "gaia")
	is.NoErr(err)

	session.Values["key"] = "somehting_else"

	recorder := httptest.NewRecorder()
	err = session.Save(req, recorder)
	is.NoErr(err)

	req.AddCookie(recorder.Result().Cookies()[0])

	client := authServer.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, http.StatusUnauthorized)
}

func TestInvalidTokenFail(t *testing.T) {
	is := is.New(t)

	config := getServerConfig()
	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))

	authServer := httptest.NewServer(addRoutes(store, config))
	defer authServer.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/secret/page.html", authServer.URL), nil)
	is.NoErr(err)

	session, err := store.Get(req, "gaia")
	is.NoErr(err)

	session.Values["token"] = "invalid token" //This token is malformed and will trigger parse error

	recorder := httptest.NewRecorder()
	err = session.Save(req, recorder)
	is.NoErr(err)

	req.AddCookie(recorder.Result().Cookies()[0])

	client := authServer.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	is.Equal(resp.StatusCode, http.StatusInternalServerError)
}

func TestProxyIntegrationIndex(t *testing.T) {

	is := is.New(t)

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Hello world")
	}))
	defer source.Close()

	config := getServerConfig()
	config.FRONTEND_SERVER = source.URL
	proxy := httptest.NewServer(NewServer(config))
	defer proxy.Close()

	r, err := http.Get(fmt.Sprintf("%v/", proxy.URL))
	is.NoErr(err)
	is.NoErr(err)
	is.Equal(r.StatusCode, http.StatusOK)
}

// This is closer to an integration test, but it hits all aspects.
func TestOnboardingIntegration(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)
	gob.Register(mitidUser{})
	u1 := database.User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}
	_, err := db.CreateUser(u1)
	is.NoErr(err)

	l, err := net.Listen("tcp", "127.0.0.1:3010")
	is.NoErr(err)

	crm := httptest.NewUnstartedServer(server.NewServer(db))
	crm.Listener.Close()
	crm.Listener = l
	crm.Start()
	defer crm.Close()

	config := getServerConfig()

	store := sessions.NewCookieStore([]byte(config.SESSION_KEY))
	ts := httptest.NewServer(addRoutes(store, config))
	defer ts.Close()

	form := url.Values{}
	form.Add("address", u1.Address)
	form.Add("darid", u1.DarId)

	req, err := http.NewRequest("POST", fmt.Sprintf("%v/account/onboarding", ts.URL), strings.NewReader(form.Encode()))
	is.NoErr(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	session, err := store.Get(req, "gaia")
	is.NoErr(err)

	session.Values["mitid"] = mitidUser{
		MitIdUUID: uuid.New().String(),
		Name:      u1.Name,
	}

	recorder := httptest.NewRecorder()
	err = session.Save(req, recorder)
	is.NoErr(err)

	req.AddCookie(recorder.Result().Cookies()[0])

	cookie := http.Cookie{
		Name:     "gaia_n",
		Value:    u1.Name,
		MaxAge:   3600,
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
	}

	req.AddCookie(&cookie)

	client := ts.Client()
	resp, err := client.Do(req)
	is.NoErr(err)
	c := resp.Cookies()
	req, err = http.NewRequest("GET", fmt.Sprintf("%v/account/onboarding", ts.URL), nil)
	is.NoErr(err)
	req.AddCookie(c[0])
	sess, err := store.Get(req, "gaia")
	is.NoErr(err)

	tokenString := sess.Values["token"].(string)
	token, err := jwt.ParseWithClaims(tokenString, &tokens.UserToken{}, func(token *jwt.Token) (any, error) {
		return []byte(config.TOKEN_SIGN_KEY), nil

	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	is.NoErr(err)

	claims, ok := token.Claims.(*tokens.UserToken)
	is.Equal(ok, true)
	is.Equal(claims.Audience, jwt.ClaimStrings{"crm", "data", "invoice"})
	is.Equal(claims.Scope, "crm:write data:read invoice:read")
}

func getServerConfig() Config {
	return Config{
		MITID_CLIENT_ID:     "0a775a87-878c-4b83-abe3-ee29c720c3e7",
		MITID_CLIENT_SECRET: "rnlguc7CM/wmGSti4KCgCkWBQnfslYr0lMDZeIFsCJweROTROy2ajEigEaPQFl76Py6AVWnhYofl/0oiSAgdtg==", //from Signaturgruppen pp env
		ENVIRONMENT:         "dev",
		TOKEN_SIGN_KEY:      "secrettokenkey",
		SESSION_KEY:         "secretsessionkey",
		FRONTEND_SERVER:     "http://localhost:3000",
	}
}

func cleanup() {
	err := os.Remove(testdb)
	if err != nil {
		panic(err)
	}
}
