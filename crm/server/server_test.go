package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"henrikkorsgaard.dk/gaia/crm/database"
)

var testdb = "test.db"

func TestGetUser(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)
	id := uuid.New().String()
	u1 := database.User{
		GaiaId:  id,
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}
	err := db.CreateUser(u1)
	is.NoErr(err)

	ts := httptest.NewServer(addRoutes(db))
	defer ts.Close()

	client := ts.Client()
	r, err := client.Get(fmt.Sprintf("%v/user/%s", ts.URL, id))
	is.NoErr(err)

	var u2 database.User
	json.NewDecoder(r.Body).Decode(&u2)

	is.Equal(u1.GaiaId, u2.GaiaId)
	is.Equal(u1.Name, u2.Name)
}

func TestCreateUser(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)
	ts := httptest.NewServer(addRoutes(db))
	defer ts.Close()
	client := ts.Client()

	var data = `{"name":"Bruno Latour", "address": "Landgreven 10, 1301 København K", "dar_id": "0a3f507a-b2e6-32b8-e044-0003ba298018"}`
	_, err := client.Post(fmt.Sprintf("%v/user", ts.URL), "application/json", strings.NewReader(data))
	is.NoErr(err)

	users, err := db.GetUsers()
	is.NoErr(err)
	is.Equal("Bruno Latour", users[0].Name)
}

func TestUpdateUser(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)

	id := uuid.New().String()
	u1 := database.User{
		GaiaId:  id,
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	err := db.CreateUser(u1)
	is.NoErr(err)

	ts := httptest.NewServer(addRoutes(db))
	defer ts.Close()
	client := ts.Client()

	var data = `{"gaia_id":"` + id + `", "name":"Bruno Latour", "address": "Constantin Hansens Gade 12, 1799 København V", "dar_id": "45380a0c-9ad1-4370-84d2-50fc574b2063"}`
	req, err := http.NewRequest("PUT", fmt.Sprintf("%v/user/%s", ts.URL, id), strings.NewReader(data))
	is.NoErr(err)

	r, err := client.Do(req)
	is.NoErr(err)
	is.Equal(r.StatusCode, http.StatusOK)

	users, err := db.GetUsers()
	is.NoErr(err)
	is.Equal("Constantin Hansens Gade 12, 1799 København V", users[0].Address)
}

func TestDeleteUser(t *testing.T) {
	defer cleanup()
	is := is.New(t)

	db := database.New(testdb)
	id := uuid.New().String()
	u1 := database.User{
		GaiaId:  id,
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}
	err := db.CreateUser(u1)
	is.NoErr(err)

	ts := httptest.NewServer(addRoutes(db))
	defer ts.Close()

	client := ts.Client()
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/user/%s", ts.URL, id), nil)
	is.NoErr(err)

	r, err := client.Do(req)
	is.NoErr(err)

	is.Equal(r.StatusCode, http.StatusOK)

	user, err := db.GetUserById(id)
	is.NoErr(err)
	is.Equal(user, database.User{})

}

func cleanup() {
	fmt.Println("Removing test database")
	err := os.Remove(testdb)
	if err != nil {
		panic(err)
	}
}
