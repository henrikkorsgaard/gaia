package server

import (
	"encoding/json"
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
)

func handleUserWithId(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			id := r.PathValue("id")

			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if id != "" && r.Method == http.MethodGet {

				user, err := db.GetUserById(id)
				if err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(user)
			}

			if id != "" && r.Method == http.MethodDelete {
				err := db.DeleteUser(id)
				if err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusOK)
				w.Write(nil)
			}

			//if method is get
			//if method is put
			//if method is post

			//fmt.Println(id)

			//TODO: Set headers globally with a proxy handler

		},
	)
}

func handleUser(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodPost {

				var user database.User
				json.NewDecoder(r.Body).Decode(&user)

				err := db.UpsertUser(user)
				if err != nil {
					panic(err)
				}
				w.WriteHeader(http.StatusCreated)
				w.Write(nil)
			}

			if r.Method == http.MethodGet {
				/*
					user, err := db.GetUserById(id)
					if err != nil {
						panic(err)
					}
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(user)
				*/
			}

			//users, err := db.GetUsers()

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}
