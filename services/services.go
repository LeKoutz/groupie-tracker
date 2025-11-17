package services

import (
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"sort"
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

// This function modifies the relations object in place.
func ProcessRelations(relations *models.Relations) {
	formatLocations(relations)
	sortDatesInLocations(relations)
	sortLocationsByDate(relations)
}

// sortDatesInLocations sorts the date arrays within each location in descending order
// (newest dates first). The sorting is done in-place, modifying the original slices.
//
// Example:
//   Input:  {"Paris": ["01-01-2024", "15-03-2024", "10-02-2024"]}
//   Output: {"Paris": ["15-03-2024", "10-02-2024", "01-01-2024"]}
func sortDatesInLocations(relations *models.Relations) {
	for location, dates := range relations.DatesLocations {
		sort.Slice(dates, func(i, j int) bool {
			return parseDate(dates[i]).After(parseDate(dates[j]))
		})
	}
}

// sortLocationsByDate sorts locations by their most recent (first) date and stores
// the sorted location names in relations.SortedLocations. Locations are sorted in
// descending order (newest first).
//
// Locations with empty date arrays are excluded from the sorted results.
//
// Example:
//   DatesLocations: {
//     "Paris, France": ["15-03-2024", ...],
//     "London, UK": ["20-03-2024", ...],
//     "Berlin, Germany": ["10-03-2024", ...]
//   }
//   Result: SortedLocations = ["London, UK", "Paris, France", "Berlin, Germany"]
func sortLocationsByDate(relations *models.Relations) {
	// Pre-allocate slice with capacity to avoid resizing
	locations := make([]string, 0, len(relations.DatesLocations))
	
	// Collect all locations that have at least one date
	for loc := range relations.DatesLocations {
		if len(relations.DatesLocations[loc]) > 0 {
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