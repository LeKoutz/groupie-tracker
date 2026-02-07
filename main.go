package main

import (
	"groupie-tracker/services"
	"groupie-tracker/models"
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
			log.Printf("\nData loaded: %d artists\nErrors: %v", len(api.All_Artists), err)
			api.SetLoadingStatus(false, true, false)
			PrecacheAllLocations(api.All_Relations)
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
	mux.HandleFunc("/api/search", handlers.SearchHandler)

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

func PrecacheAllLocations(relations []models.Relations) {
	// map to ensure we geocode each location once
	uniqueLocs := make(map[string]bool)
	
	for _, rel := range relations {
		for locName := range rel.DatesLocations {
			uniqueLocs[locName] = true
		}
	}
	// convert map to slice
	var allLocs []string
	for locName := range uniqueLocs {
		allLocs = append(allLocs, locName)
	}
	log.Printf("Starting location geocoding for %d locations", len(allLocs))
	go services.Geocode(allLocs)
	log.Println("Background geocoding complete")
	}