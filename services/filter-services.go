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

// FilterArtists filters artists based on the provided parameters.
func FilterArtists(artists []models.Artists, locations []models.Locations, filters models.FilterParameters) []models.Artists {
	// Optimization: Index locations by Artist ID for O(1) lookup.
	// This avoids looping through the locations array for every single artist.
	locMap := make(map[int][]string, len(locations))
	for _, l := range locations {
		locMap[l.ID] = l.Locations
	}
	var filtered []models.Artists
	for _, artist := range artists {
		// Pass the specific locations for this artist to the matcher
		if matchesFilters(artist, locMap[artist.ID], filters) {
			filtered = append(filtered, artist)
		}
	}
	return filtered
}

// matchesFilters checks if a single artist satisfies all filter criteria.
func matchesFilters(artist models.Artists, artistLocs []string, f models.FilterParameters) bool {
	// 1. Creation Date
	if artist.CreationDate < f.MinCreationDate || artist.CreationDate > f.MaxCreationDate {
		return false
	}
	// 2. First Album Year
	year := ExtractYearFromDate(artist.FirstAlbum)
	if year < f.MinFirstAlbumYear || year > f.MaxFirstAlbumYear {
		return false
	}
	// 3. Members
	if len(artist.Members) < f.MinMembers || len(artist.Members) > f.MaxMembers {
		return false
	}
	// 4. Locations
	// If any locations are selected, the artist must match at least one.
	if len(f.SelectedLocations) > 0 {
		return hasMatchingLocation(artistLocs, f.SelectedLocations)
	}
	return true
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
