package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/henrikkorsgaard/gaia/auth/server"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	config := server.Config{}
	err := env.Parse(&config)
	if err != nil {
		panic(fmt.Sprintf("unable to parse ennvironment variables: %e", err))
	}

	log.Fatal(http.ListenAndServe(":3020", server.NewServer(config)))
}
