package services

import (
	"groupie-tracker/models"
	"sort"
	"strings"
)

// FilterArtists filters artists based on the provided parameters.
func FilterArtists(artists []models.Artists, locations []models.Locations, filters models.FilterParameters) []models.Artists {
	locMap := make(map[int][]string, len(locations))
	for _, l := range locations {
		locMap[l.ID] = l.Locations
	}
	var filtered []models.Artists
	for _, artist := range artists {
		if matchesFilters(artist, locMap[artist.ID], filters) {
			filtered = append(filtered, artist)
		}
	}
	return filtered
}

// matchesFilters checks if a single artist satisfies all filter criteria.
func matchesFilters(artist models.Artists, artistLocs []string, f models.FilterParameters) bool {
	if !inRange(artist.CreationDate, f.MinCreationDate, f.MaxCreationDate) {
		return false
	}
	year := ExtractYearFromDate(artist.FirstAlbum)
	if !inRange(year, f.MinFirstAlbumYear, f.MaxFirstAlbumYear) {
		return false
	}
	if !inRange(len(artist.Members), f.MinMembers, f.MaxMembers) {
		return false
	}
	if len(f.SelectedLocations) > 0 {
		return hasMatchingLocation(artistLocs, f.SelectedLocations)
	}
	return true
}

// ExtractYearFromDate extracts the year from a date string.
func ExtractYearFromDate(dateStr string) int {
	t, err := parseDate(dateStr)
	if err != nil {
		return 0
	}
	return t.Year()
}

// hasMatchingLocation handles the hierarchical location check.
func hasMatchingLocation(artistLocs []string, selected []string) bool {
	for _, s := range selected {
		target := strings.ToLower(s)
		for _, loc := range artistLocs {
			formatted := strings.ToLower(formatLocationName(loc))
			if strings.Contains(formatted, target) {
				return true
			}
		}
	}
	return false
}

// inRange checks if a value is within the specified range (inclusive).
func inRange(val, min, max int) bool {
	return val >= min && val <= max
}

// ParseLocations collects all unique, formatted locations from the dataset.
func ParseLocations(locationData []models.Locations) []string {
	unique := make(map[string]bool)
	for _, data := range locationData {
		for _, loc := range data.Locations {
			formatted := formatLocationName(loc)
			unique[formatted] = true
		}
	}
	result := make([]string, 0, len(unique))
	for loc := range unique {
		result = append(result, loc)
	}
	sort.Strings(result)
	return result
}
