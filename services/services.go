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
// It accepts a few common separator variants and trims whitespace. On failure it
// returns a non-nil error so callers can decide how to handle invalid dates.
func parseDate(dateStr string) (time.Time, error) {
	s := strings.TrimSpace(dateStr)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// remove leading '*' markers that appear in the API and trim spaces
	s = strings.TrimLeftFunc(s, func(r rune) bool { return r == '*' || unicode.IsSpace(r) })

	// normalize common separators to '-'
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, ".", "-")

	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q", dateStr)
	}
	return t, nil
}

// titleCase converts a string into Title Case for each word while trimming
// excessive whitespace.
func titleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// formatLocationName converts "city-country" into "City, Country".
func formatLocationName(loc string) string {
	loc = strings.ReplaceAll(loc, "_", " ")
	// Split by the last hyphen so values like "san-juan-puerto-rico" are
	// interpreted sensibly as "san-juan, puerto-rico" rather than
	// "san, juan, puerto, rico". If there's no hyphen, title-case the whole
	// string.
	if idx := strings.LastIndex(loc, "-"); idx != -1 {
		left := strings.ReplaceAll(loc[:idx], "-", " ")
		right := strings.ReplaceAll(loc[idx+1:], "-", " ")
		left = titleCase(left)
		right = titleCase(right)
		return strings.TrimSpace(left) + ", " + strings.TrimSpace(right)
	}
	return titleCase(loc)
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

// dateNewer returns true if dateA is newer (later) than dateB.
// It returns false if either date cannot be parsed (treating unparseable dates as older).
func dateNewer(dateA, dateB string) bool {
	a, errA := parseDate(dateA)
	b, errB := parseDate(dateB)
	if errA != nil || errB != nil {
		return errA == nil
	}
	return a.After(b)
}

// sortDatesInLocations sorts the date arrays within each location in descending order
// (newest dates first). The sorting is done in-place.
func sortDatesInLocations(relations *models.Relations) {
	for loc, dates := range relations.DatesLocations {
		if len(dates) <= 1 {
			continue
		}
		// Use centralized comparison helper to avoid repeating parse/err handling here.
		sort.SliceStable(dates, func(i, j int) bool { return dateNewer(dates[i], dates[j]) })
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
		dateI, _ := parseDate(relations.DatesLocations[locations[i]][0])
		dateJ, _ := parseDate(relations.DatesLocations[locations[j]][0])
		return dateI.After(dateJ)
	})
	relations.SortedLocations = locations
}
