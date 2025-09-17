package main

import (
	"log"
	"net/http"
	"os"

	"travel-search/handlers"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	http.HandleFunc("/search", handlers.SearchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
