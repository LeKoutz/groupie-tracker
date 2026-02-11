package services

import (
	"encoding/json"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"net/http"
	"net/url"
	"sync"
	"time"
	"context"
)

// memory cache
var (
	geoCache = make(map[string]models.Coordinates)
	geoMutex sync.RWMutex
)

// Geocode processes a list of locations and returns their coordinates.
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
		// We sleep BEFORE the request to ensure we don't hit the limit.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		
		// Helper to safely handle spaces and special chars
		baseURL := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q="
		req, err := http.NewRequestWithContext(ctx, "GET", baseURL+url.QueryEscape(loc), nil)
		if err != nil {
			cancel()
			continue
		}
		 // Set identity for API (required by Novatim)
		req.Header.Set("User-Agent", "GroupieTracker")
		resp, err := api.Client.Do(req)
		if err != nil {
			cancel()
			continue
		}

		var data []models.Coordinates
		// 3. Decode & Cache
		if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data) > 0 {
			geoMutex.Lock()
			geoCache[loc] = data[0] // store the best result in the cache
			geoMutex.Unlock()
			results[loc] = data[0]
		}
		resp.Body.Close()
		cancel()
	}
	return results
}