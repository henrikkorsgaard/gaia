package server

import (
	"net/http"

	"henrikkorsgaard.dk/gaia/crm/database"
)

// TODO: Return CRM app view
func viewHandler(db *database.UserDatabase) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}
