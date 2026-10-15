package main

import (
	"log"
	"net/http"
)

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("marketplace domain service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler(client)))
}
