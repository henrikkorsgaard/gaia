package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/henrikkorsgaard/gaia/auth/server"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	fmt.Println("Server is running on port 3020...")
	// https://github.com/gorilla/securecookie
	config := server.Config{
		MITID_CLIENT_ID:     os.Getenv("MITID_CLIENT_ID"),
		MITID_CLIENT_SECRET: os.Getenv("MITID_CLIENT_SECRET"),
		ENVIRONMENT:         os.Getenv("ENVIRONMENT"),
		TOKEN_SIGN_KEY:      os.Getenv("TOKEN_SIGN_KEY"),
		SESSION_KEY:         os.Getenv("SESSION_KEY"),
		FRONTEND_SERVER:     os.Getenv("FRONTEND_SERVER"),
	}

	log.Fatal(http.ListenAndServe(":3020", server.NewServer(config)))
}
