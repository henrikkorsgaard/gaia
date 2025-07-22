package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
	"henrikkorsgaard.dk/gaia/crm/tokens"
)

func handleMatch(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method == http.MethodPost {
				var userRequest database.User
				json.NewDecoder(r.Body).Decode(&userRequest)

				// guard function should return
				if userRequest.DarId != "" {
					fmt.Println("we want to try a match on address")
				}

				/*
					if we cannot match a user, there is no reason to create a token

				*/

				if userRequest.MitIdUUID != "" {
					user, err := db.GetUserMitIDUUID(userRequest.MitIdUUID)
					if err != nil {
						//return 404 from db or here!
						fmt.Println(err)
						fmt.Println("Fuck that day I need to go through all error handling")
					}

					token, err := tokens.NewUserToken(user)
					if err != nil {

						fmt.Println("handle it")
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(token))
				}
			}
		},
	)
}
