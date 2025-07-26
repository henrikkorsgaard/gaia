package server

import (
	"encoding/gob"
	"fmt"
	"io"
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
	app "github.com/henrikkorsgaard/gaia/app/server"
	"github.com/henrikkorsgaard/gaia/auth/tokens"
	"github.com/henrikkorsgaard/gaia/crm/database"
	"github.com/henrikkorsgaard/gaia/crm/server"
	"github.com/matryer/is"
)

var testdb = "test.db"

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

func TestProxyIntegrationIndex(t *testing.T) {

	is := is.New(t)

	l, err := net.Listen("tcp", "localhost:3020")
	is.NoErr(err)

	config := getServerConfig()

	prxy := httptest.NewUnstartedServer(NewServer(config))
	prxy.Listener.Close()
	prxy.Listener = l
	prxy.Start()
	defer prxy.Close()

	fl, err := net.Listen("tcp", "localhost:3000")
	is.NoErr(err)

	frontend := httptest.NewUnstartedServer(app.NewServer("../../app/static"))
	frontend.Listener.Close()
	frontend.Listener = fl
	frontend.Start()
	defer frontend.Close()

	urrl := fmt.Sprintf("%v/", frontend.URL)
	fmt.Println(urrl)
	r, err := http.Get(urrl)
	is.NoErr(err)
	body, err := io.ReadAll(r.Body)
	is.NoErr(err)
	is.Equal(strings.Contains(string(body), "<h1>Hello World</h1>"), true)
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
