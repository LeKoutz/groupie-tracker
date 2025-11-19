package services

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"groupie-tracker/api"
	"groupie-tracker/models"
)

func GetArtistByID(id int) (*models.Artists, error) {
	for i := range api.All_Artists {
		if api.All_Artists[i].ID == id {
			return &api.All_Artists[i], nil
		}
	}
	return nil, fmt.Errorf("Error: Artist ID %d not found", id)
}

func GetLocationsByID(id int) (*models.Locations, error) {
	for i := range api.All_Locations {
		if api.All_Locations[i].ID == id {
			return &api.All_Locations[i], nil
		}
	}
	return nil, fmt.Errorf("Error: No locations found for ID %d", id)
}

func GetDatesByID(id int) (*models.Dates, error) {
	for i := range api.All_Dates {
		if api.All_Dates[i].ID == id {
			return &api.All_Dates[i], nil
		}
	}
	return nil, fmt.Errorf("Error: No dates found for ID %d", id)
}

func GetRelationsByID(id int) (*models.Relations, error) {
	for i := range api.All_Relations {
		if api.All_Relations[i].ID == id {
			relations := &api.All_Relations[i]
			ProcessRelations(relations)
			return relations, nil
		}
	}
	return nil, fmt.Errorf("Error: No relations found for ID %d", id)
}

const DateFormat = "02-01-2006" // dd-mm-yyyy

// parseDate parses a date string in the format "dd-mm-yyyy" and returns a time.Time.
// On parse error it returns the zero time.
func parseDate(dateStr string) time.Time {
	t, _ := time.Parse(DateFormat, dateStr)
	return t
}

// capitalize makes the first rune uppercase and the rest lowercase.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(strings.ToLower(strings.TrimSpace(s)))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// formatLocationName converts "city-country" into "City, Country".
func formatLocationName(loc string) string {
	loc = strings.ReplaceAll(loc, "_", " ")
	parts := strings.Split(loc, "-")
	for i, p := range parts {
		parts[i] = capitalize(p)
	}
	return strings.Join(parts, ", ")
}

// formatLocations replaces the keys in DatesLocations with formatted names.
func formatLocations(relations *models.Relations) {
	formatted := make(map[string][]string, len(relations.DatesLocations))
	for loc, dates := range relations.DatesLocations {
		formatted[formatLocationName(loc)] = dates
	}
	relations.DatesLocations = formatted
}

// This function modifies the relations object in place.
func ProcessRelations(relations *models.Relations) {
	formatLocations(relations)
	sortDatesInLocations(relations)
	sortLocationsByDate(relations)
}

// sortDatesInLocations sorts the date arrays within each location in descending order
// (newest dates first). The sorting is done in-place.
func sortDatesInLocations(relations *models.Relations) {
	for loc, dates := range relations.DatesLocations {
		if len(dates) <= 1 {
			continue
		}
		// sorting the slice referenced by the map entry; the slice header is copied
		// but refers to the same backing array, so modifications are visible in the map.
		sort.Slice(dates, func(i, j int) bool {
			return parseDate(dates[i]).After(parseDate(dates[j]))
		})
		relations.DatesLocations[loc] = dates
	}
}

// sortLocationsByDate sorts locations by their most recent (first) date and stores
// the sorted location names in relations.SortedLocations. Locations are sorted in
// descending order (newest first). Locations with no dates are excluded.
func sortLocationsByDate(relations *models.Relations) {
	// Pre-allocate slice with capacity to avoid resizing
	locations := make([]string, 0, len(relations.DatesLocations))
	for loc, dates := range relations.DatesLocations {
		if len(dates) > 0 {
			locations = append(locations, loc)
		}
	}

	// Sort locations by their most recent date (index 0)
	sort.Slice(locations, func(i, j int) bool {
		dateI := parseDate(relations.DatesLocations[locations[i]][0])
		dateJ := parseDate(relations.DatesLocations[locations[j]][0])
		return dateI.After(dateJ)
	})

	relations.SortedLocations = locations
}
