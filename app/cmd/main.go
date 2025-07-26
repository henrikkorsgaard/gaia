package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/henrikkorsgaard/gaia/app/server"
)

func main() {
	fmt.Println("Server is running on port 3000...")
	log.Fatal(http.ListenAndServe(":3000", server.NewServer("static")))
}
