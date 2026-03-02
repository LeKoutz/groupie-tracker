package services

import (
	"groupie-tracker/models"
	"sort"
	"strings"
)

// FilterArtists filters artists based on the provided parameters
func FilterArtists(artists []models.Artists, locations []models.Locations, f models.FilterParameters) []models.Artists {
	locMap := make(map[int][]string)
	for _, l := range locations {
		locMap[l.ID] = l.Locations
	}
	var filtered []models.Artists
	for _, a := range artists {
		// Single-pass filtering with continue statements
		if !inRange(a.CreationDate, f.MinCreationDate, f.MaxCreationDate) ||
			!inRange(extractYear(a.FirstAlbum), f.MinFirstAlbumYear, f.MaxFirstAlbumYear) ||
			!inRange(len(a.Members), f.MinMembers, f.MaxMembers) ||
			(len(f.SelectedLocations) > 0 && !hasLocation(locMap[a.ID], f.SelectedLocations)) {
			continue
		}
		filtered = append(filtered, a)
	}
	return filtered
}

// extractYear extracts the year from a date string
func extractYear(dateStr string) int {
	if t, err := parseDate(dateStr); err == nil {
		return t.Year()
	}
	return 0
}

// hasLocation checks if an artist's location matches any of the selected locations
func hasLocation(artistLocs []string, selected []string) bool {
	for _, s := range selected {
		lower := strings.ToLower(s)
		for _, loc := range artistLocs {
			if strings.Contains(strings.ToLower(formatLocationName(loc)), lower) {
				return true
			}
		}
	}
	return false
}

// inRange checks if a value is within the specified range (inclusive)
func inRange(val, min, max int) bool {
	return val >= min && val <= max
}

// ParseLocations collects all unique, formatted locations.
func ParseLocations(locations []models.Locations) []string {
	unique := make(map[string]bool)
	for _, l := range locations {
		for _, loc := range l.Locations {
			unique[formatLocationName(loc)] = true
		}
	}
	result := make([]string, 0, len(unique))
	for loc := range unique {
		result = append(result, loc)
	}
	sort.Strings(result)
	return result
}
