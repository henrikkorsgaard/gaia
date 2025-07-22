package server

import (
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
)

func handleMatch(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//if there is a match on mitiduuid, then we return access_token with gaia id + scope + aud [crm]
			//aud:
			return
		},
	)
}
