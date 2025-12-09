package main

import (
	"groupie-tracker/api"
	"groupie-tracker/handlers"
	"log"
	"net/http"
	"os"
)

func main () {
	api.SetLoadingStatus(true, false, false)
	// Initialize the data structures
	go func() {
		err := api.InitializeData()
		if err != nil {
			log.Printf("Failed to load data with error: %v", err)
			api.SetLoadingStatus(false, false, true)
		} else {
			log.Printf("\nData loaded: %d artists, %d locations, %d dates, %d relations\nErrors: %v", len(api.All_Artists), len(api.All_Locations), len(api.All_Dates), len(api.All_Relations), err)
			api.SetLoadingStatus(false, true, false)
		}
	}()
	// Refresh the data occasionally
	go api.RefreshData()
	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/artist/", handlers.ArtistDetailsHandler)
	mux.HandleFunc("/loading/", handlers.LoadingHandler)
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
