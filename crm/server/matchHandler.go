package server

import (
	"encoding/json"
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
	"henrikkorsgaard.dk/gaia/crm/tokens"
)

var (
	ErrMatchMissingMitIdUUID = "error: missing mitid_uuid. unable to match without mitid_uuid."
	ErrMatchMissingDarId     = "error: missing dar_id. unable to match based on address without dar_id."
)

func handleMatch(db *database.UserDatabase) http.Handler {
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

				user, err := db.GetUserMitIDUUID(userRequest.MitIdUUID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Identity match requires GaiaId
				// If not succesful with mitid_uuid, we try dar_id
				if user.GaiaId == "" {

					if userRequest.DarId == "" {
						//400: Missing dar_id
						http.Error(w, ErrMatchMissingDarId, http.StatusBadRequest)
						return
					}

					/*
						Attention: In a real system, we would try to match the user on the dar_id
						In this demo we simply create the user based on dar_id
					*/
					user, err = db.CreateUser(userRequest)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}

				token, err := tokens.NewUserToken(user)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(token))
				return
			}
		},
	)
}
