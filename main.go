package main

import (
	"RateLimit/api"
	"log"
)

func main() {
	log.Print("Starting...")

	if err := api.Serve(); err != nil {
		log.Fatalf("Failed to run Seerve. Err '%v'", err)
	}
}
