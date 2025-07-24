package server

import (
	"encoding/json"
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
	"henrikkorsgaard.dk/gaia/crm/tokens"
)

var (
	ErrMatchMissingMitIdUUID = "error: missing mitid_uuid. unable to match without mitid_uuid."
)

/*
	matchHandler will return a token for the user with the claims
	OR 404 if unable to match
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

				// Check if any match was succesful and return token
				if user.GaiaId != "" {
					token, err := tokens.NewUserToken(user)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)

						return
					}

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(token))

					return
				}

				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("unable to match identity"))

				return
			}
		},
	)
}
