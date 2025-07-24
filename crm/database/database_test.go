package database

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

var testdb = "test.db"

func TestCreateUser(t *testing.T) {
	defer cleanup()

	is := is.New(t)
	db := New(testdb)
	u1 := User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	u2, err := db.CreateUser(u1)
	is.NoErr(err)
	is.Equal(u2.GaiaId != "", true)
}

func TestGetUser(t *testing.T) {
	defer cleanup()

	is := is.New(t)
	db := New(testdb)

	id := uuid.New().String()
	u1 := User{
		GaiaId:  id,
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	_, err := db.CreateUser(u1)
	is.NoErr(err)

	u2, err := db.GetUserById(id)
	is.NoErr(err)
	is.Equal(id, u2.GaiaId)
	is.Equal(u1.Name, u2.Name)
}

func TestBulkCreate(t *testing.T) {

	is := is.New(t)
	db := New(testdb)
	u1 := User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	u2 := User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	u3 := User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	u4 := User{
		Name:    "Bruno Latour",
		Address: "Landgreven 10, 1301 København K",
		DarId:   "0a3f507a-b2e6-32b8-e044-0003ba298018",
	}

	rows, users, err := db.BulkCreateUsers([]User{u1, u2, u3, u4})
	is.NoErr(err)
	is.Equal(rows, int64(4))
	for _, u := range users {
		is.Equal(u.GaiaId != "", true)
	}
}

func TestGetUsers(t *testing.T) {
	defer cleanup()
	is := is.New(t)
	db := New(testdb)
	users, err := db.GetUsers()
	is.NoErr(err)
	is.Equal(len(users), 4)
}

func cleanup() {
	err := os.Remove(testdb)
	if err != nil {
		panic(err)
	}
}
