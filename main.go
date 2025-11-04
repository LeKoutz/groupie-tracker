package main

import (
	"groupie-tracker/api"
	"groupie-tracker/handlers"
	"log"
	"net/http"
	"os"
)

func main () {
	// Initialize the data structures
	err := api.InitializeData()
	if err != nil {
		log.Printf("Failed to load data with error: %v", err)
	}
	log.Printf("\nData loaded: %d artists, %d locations, %d dates, %d relations\nErrors: %v", len(api.All_Artists), len(api.All_Locations), len(api.All_Dates), len(api.All_Relations), err)
	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/artist/", handlers.ArtistDetailsHandler)
	mux.HandleFunc("/static/", handlers.ResourcesHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Server starting on: http://localhost:" + port)
	log.Println("Press CTRL+C to stop the server")
	log.Fatal(http.ListenAndServe(addr, mux))
}
