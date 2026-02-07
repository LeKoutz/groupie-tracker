package services

import (
	"sync"
	"groupie-tracker/models"
	"groupie-tracker/api"
	"net/http"
	"encoding/json"
	"net/url"
)

// memory cache
var (
	geoCache = make(map[string]models.Coordinates)
	geoMutex sync.RWMutex
)

// Geocode processes a list of locations and returns their coordinates.
// It reuses the existing api.Client for timeouts and efficiency.
func Geocode(locations []string) map[string]models.Coordinates {
	results := make(map[string]models.Coordinates)
	 
		for _, loc := range locations {
		// 1. Check Cache
		geoMutex.RLock()
		val, cached := geoCache[loc]
		geoMutex.RUnlock()
		if cached {
			results[loc] = val
			continue
		}

		// 2. Fetch from API
		// We use url.QueryEscape to safely handle spaces and special chars
		baseURL := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" // This helper converts it to New+York or New%20York so the server understands it + we limit the results to the top one
		req, _ := http.NewRequest("GET", baseURL+url.QueryEscape(loc), nil)
		req.Header.Set("User-Agent", "GroupieTracker") // set identity for api
		resp, err := api.Client.Do(req)
		if err == nil {
			var data []models.Coordinates
			// 3. Decode & Cache
			if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data) > 0 { // parse the response
				geoMutex.Lock()
				geoCache[loc] = data[0] // store the best result in the cache
				geoMutex.Unlock()
				results[loc] = data[0]
			}
			resp.Body.Close() // prevent memory leaks
		}
	}
	return results
}