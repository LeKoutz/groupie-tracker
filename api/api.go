package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/models"
	"net/http"
)

const (
	ARTISTS_API		= "https://groupietrackers.herokuapp.com/api/artists"
	LOCATIONS_API	= "https://groupietrackers.herokuapp.com/api/locations"
	DATES_API		= "https://groupietrackers.herokuapp.com/api/dates"
	RELATIONS_API	= "https://groupietrackers.herokuapp.com/api/relations"
)

func FetchArtists() ([]models.Artists, error) {
	resp, err := http.Get(ARTISTS_API)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s with error: %v", ARTISTS_API, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}

	var artists []models.Artists
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return artists, nil
}

func FetchLocations() ([]models.Locations, error) {
	resp, err := http.Get(LOCATIONS_API)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s with error: %v", LOCATIONS_API, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}

	var concert_locations models.LocationsIndex
	if err := json.NewDecoder(resp.Body).Decode(&concert_locations); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return concert_locations.Index, nil
}

func FetchDates() ([]models.Dates, error) {
	resp, err := http.Get(DATES_API)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s with error: %v", DATES_API, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}
	var concert_dates models.DatesIndex
	if err := json.NewDecoder(resp.Body).Decode(&concert_dates); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return concert_dates.Index, nil
}

func FetchRelations() ([]models.Relations, error) {
	resp, err := http.Get(RELATIONS_API)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from %s with error: %v", RELATIONS_API, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}
	var relations models.RelationIndex
	if err := json.NewDecoder(resp.Body).Decode(&relations); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return relations.Index, nil
}
