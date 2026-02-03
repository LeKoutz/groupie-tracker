package services

import (
	"groupie-tracker/models"
	"sort"
)

func ExtractYearFromDate(dateStr string) int {
	t, err := parseDate(dateStr)
	if err != nil {
		return 0
	}
	return t.Year()
}

// ParseLocations collects all unique, formatted locations from the dataset.
func ParseLocations(locationData []models.Locations) []string {
	// 1. Use a map to collect unique values (Set)
	unique := make(map[string]bool)

	for _, data := range locationData {
		for _, loc := range data.Locations {
			// Format first to ensure "usa" and "USA" count as the same valid location
			formatted := formatLocationName(loc)
			unique[formatted] = true
		}
	}

	// 2. Extract keys into a slice
	result := make([]string, 0, len(unique))
	for loc := range unique {
		result = append(result, loc)
	}

	// 3. Sort for a clean dropdown UI
	sort.Strings(result)

	return result
}
