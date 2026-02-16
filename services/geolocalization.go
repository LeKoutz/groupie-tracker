package services

import (
	"context"
	"encoding/json"
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// memory cache
var (
	geoCache  = make(map[string]models.Coordinates)
	geoMutex  sync.RWMutex
	cacheFile = "locations.json" // to store cached locations
)

// InitGeoCache only loads the file. Background filling is started separately.
func InitGeoCache() {
	loadCache()
}

func loadCache() {
	file, err := os.Open(cacheFile)
	if err != nil {
		fmt.Println("No cache file found, starting with empty cache.")
		return
	}
	defer file.Close()

	geoMutex.Lock()
	defer geoMutex.Unlock()
	if err := json.NewDecoder(file).Decode(&geoCache); err == nil {
		fmt.Printf("Loaded %d locations from cache.\n", len(geoCache))
	}
}

func saveCache() {
	geoMutex.RLock()
	data, err := json.MarshalIndent(geoCache, "", "  ")
	geoMutex.RUnlock()

	if err == nil {
		// Ignore errors on save (non-critical)
		_ = os.WriteFile(cacheFile, data, 0644)
	}
}

// FillCacheBackground iterates through all relations and fetches missing coordinates.
// It uses formatLocationName to ensure keys match the frontend requests.
func FillCacheBackground() {
	fmt.Println("Starting background geocoding...")

	uniqueLocs := make(map[string]bool)

	// Collect all unique formatted locations
	for _, rel := range api.All_Relations {
		for rawLoc := range rel.DatesLocations {
			formatted := formatLocationName(rawLoc)
			uniqueLocs[formatted] = true
		}
	}

	dirty := false
	count := 0

	for loc := range uniqueLocs {
		// 1. Check Cache
		geoMutex.RLock()
		_, exists := geoCache[loc]
		geoMutex.RUnlock()

		if exists {
			continue
		}

		// 2. Fetch if missing
		coord, err := fetchSingleCoordinate(loc)
		if err == nil {
			geoMutex.Lock()
			geoCache[loc] = coord
			geoMutex.Unlock()

			dirty = true
			count++
			fmt.Printf("Cached: %s\n", loc)

			// Save periodically to prevent data loss on crash
			if count%5 == 0 {
				saveCache()
				dirty = false
			}
		} else {
			fmt.Printf("Failed to fetch %s: %v\n", loc, err)
		}
	}
	// Final save
	if dirty {
		saveCache()
	}
	fmt.Println("Geolocalization background update complete.")
}

// Geocode processes a list of locations and returns their coordinates.
func Geocode(locations []string) map[string]models.Coordinates {
	results := make(map[string]models.Coordinates)

	for _, loc := range locations {
		// Check Cache
		geoMutex.RLock()
		val, cached := geoCache[loc]
		geoMutex.RUnlock()

		if cached {
			results[loc] = val
			continue
		}

		// Fetch on demand if not in cache
		coord, err := fetchSingleCoordinate(loc)
		if err == nil {
			geoMutex.Lock()
			geoCache[loc] = coord
			geoMutex.Unlock()
			results[loc] = coord
			
			go saveCache()
		}
	}
	return results
}

func fetchSingleCoordinate(loc string) (models.Coordinates, error) {
	// Short timeout prevents hanging requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	baseURL := "https://nominatim.openstreetmap.org/search?format=json&limit=1&q="
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+url.QueryEscape(loc), nil)
	if err != nil {
		return models.Coordinates{}, err
	}
	// User identity policy by Novatim
	req.Header.Set("User-Agent", "GroupieTracker")

	resp, err := api.Client.Do(req)
	if err != nil {
		return models.Coordinates{}, err
	}
	defer resp.Body.Close()

	var data []models.Coordinates
	if json.NewDecoder(resp.Body).Decode(&data) == nil && len(data) > 0 {
		return data[0], nil
	}
	return models.Coordinates{}, fmt.Errorf("no coordinates found")
}