package main

import (
	"fmt"

	"github.com/caarlos0/env"
	"github.com/henrikkorsgaard/gaia/dmi/dmi"
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

	config := dmi.Config{}
	err = env.Parse(&config)
	if err != nil {
		panic(fmt.Sprintf("unable to parse ennvironment variables: %e", err))
	}

	fmt.Println(config.API_KEY)
}
