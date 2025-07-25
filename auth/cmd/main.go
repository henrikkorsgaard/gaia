package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/henrikkorsgaard/gaia/auth/server"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("Server is running on port 3020...")
	// https://github.com/gorilla/securecookie
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	log.Fatal(http.ListenAndServe(":3020", server.NewServer(store)))
}
