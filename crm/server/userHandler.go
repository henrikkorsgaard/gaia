package server

import (
	"encoding/json"
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
)

func userIdHandler(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			id := r.PathValue("id")

			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if id != "" && r.Method == http.MethodGet {

				user, err := db.GetUserById(id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(user)
				return
			}

			if id != "" && r.Method == http.MethodPut {

				var user database.User
				json.NewDecoder(r.Body).Decode(&user)

				err := db.UpdateUserById(user)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				return
			}

			if id != "" && r.Method == http.MethodDelete {
				err := db.DeleteUser(id)
				if err != nil {
					panic(err)
				}

				w.WriteHeader(http.StatusOK)
				return
			}

			http.Error(w, "", http.StatusMethodNotAllowed)

		},
	)
}

func userHandler(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodPost {

				var user database.User
				json.NewDecoder(r.Body).Decode(&user)

				newUser, err := db.CreateUser(user)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(newUser)
				return
			}

			if r.Method == http.MethodGet {

				users, err := db.GetUsers()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(users)
				return
			}

			http.Error(w, "", http.StatusMethodNotAllowed)

		},
	)
}
