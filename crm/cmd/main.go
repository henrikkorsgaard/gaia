package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"henrikkorsgaard.dk/gaia/crm/database"
	"henrikkorsgaard.dk/gaia/crm/server"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	dbhost := os.Getenv("DATABASE_HOST")
	port := os.Getenv("SERVER_PORT")
	db := database.New(dbhost)

	fmt.Printf("CRM Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, server.NewServer(db)))
}
