package services

import (
	"sync"
	"groupie-tracker/models"
	"groupie-tracker/api"
	"net/http"
	"encoding/json"
	"net/url"
	"context"
	"time"
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
	var wg sync.WaitGroup
	var mu sync.Mutex
	 
		for _, loc := range locations {
		// 1. Check Cache
		geoMutex.RLock()
		val, cached := geoCache[loc]
		geoMutex.RUnlock()
		if cached {
			results[loc] = val
			mu.Unlock()
			continue
		}
		// fetch in parallel if not in cache
		wg.Add(1)
		go func(loc string) {
			defer wg.Done()
			
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

		// 2. Fetch from API
		// We use url.QueryEscape to safely handle spaces and special chars
		baseURL := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q=" // This helper converts it to New+York or New%20York so the server understands it + we limit the results to the top one
		req, err := http.NewRequestWithContext(ctx, "GET", baseURL+url.QueryEscape(loc), nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "GroupieTracker") // set identity for api
		resp, err := api.Client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		var data []models.Coordinates
			// 3. Decode & Cache
		if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data) > 0 { // parse the response
			geoMutex.Lock()
			geoCache[loc] = data[0] // store the best result in the cache
			geoMutex.Unlock()
			mu.Lock()
			results[loc] = data[0]
			mu.Unlock()
			}
		}(loc)
		}
	wg.Wait() // wait for all goroutines to finish or time out
	return results
}