package server

import (
	"net/http"
)

func GetUserById(userId string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}

func GetUsers() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}

func CreateUser() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}

// TODO send a user json obj
func UpdateUser() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}

// TODO send a user json obj
func DeleteUser(userId string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			//TODO: Set headers globally with a proxy handler
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte("We need to return json"))
		},
	)
}
