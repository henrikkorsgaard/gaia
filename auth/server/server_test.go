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

// This is closer to an integration test, but it hits all aspects.
func TestOnboarding(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)
	store := sessions.NewCookieStore([]byte("Test secrets"))
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

	ts := httptest.NewServer(addRoutes(store))
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
		return []byte("tokensecret"), nil

	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	is.NoErr(err)

	claims, ok := token.Claims.(*tokens.UserToken)
	is.Equal(ok, true)
	is.Equal(claims.Audience, jwt.ClaimStrings{"crm", "data", "invoice"})
	is.Equal(claims.Scope, "crm:write data:read invoice:read")
}

func cleanup() {
	err := os.Remove(testdb)
	if err != nil {
		panic(err)
	}
}
