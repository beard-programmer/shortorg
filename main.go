package main

import (
	"log"

	"github.com/beard-programmer/shortorg/internal"
)

func main() {
	// Initialize the application
	app := new(internal.App).New()

	// Start the HTTP server
	if err := app.StartServer(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
