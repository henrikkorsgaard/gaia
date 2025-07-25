package server

import (
	"encoding/json"
	"net/http"

	"github.com/henrikkorsgaard/gaia/crm/database"
)

var (
	ErrMatchMissingMitIdUUID = "error: missing mitid_uuid. unable to match without mitid_uuid."
)

/*
	matchHandler will return a user with an id OR 404 if unable to match
*/

func matchHandler(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				var userRequest database.User
				json.NewDecoder(r.Body).Decode(&userRequest)

				/*
					We always need the MitIdUUID.
					- To match an existing user
					- OR to create a new user with address
				*/
				if userRequest.MitIdUUID == "" {
					//400: Missing mitid_uuid
					http.Error(w, ErrMatchMissingMitIdUUID, http.StatusBadRequest)
					return
				}

				// Initial Mitid match
				user, err := db.GetUserMitIDUUID(userRequest.MitIdUUID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if user.GaiaId == "" && userRequest.DarId != "" {
					user, err = db.CreateUser(userRequest)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				// Check if any match was succesful and return user
				if user.GaiaId != "" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(user)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("unable to match identity"))

				return
			}
		},
	)
}
