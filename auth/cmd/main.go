package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env"
	"github.com/henrikkorsgaard/gaia/auth/server"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		// TODO: Create config if not existing
		/*
			if errors.Is(err, fs.ErrNotExist) {
			}*/
		panic(fmt.Sprintf("unable to load .env file: %e", err))
	}

	config := server.Config{}
	err = env.Parse(&config)
	if err != nil {
		panic(fmt.Sprintf("unable to parse ennvironment variables: %e", err))
	}
	//TODO:Port should come from config as well
	log.Fatal(http.ListenAndServe(":3020", server.NewServer(config)))
}
