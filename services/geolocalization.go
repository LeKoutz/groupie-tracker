package services

import (
	"sync"
)

// memory cache
var (
	geoCache map[string]models.Coordinates
	geoMutex sync.RWMutex
)

// Geocode processes a list of locations and returns their coordinates.
// It reuses your existing api.Client for timeouts and efficiency.
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
	}
}

// rate limit policy
// time.Sleep(1 * time.Second)