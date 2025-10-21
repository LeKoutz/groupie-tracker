package api

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/models"
	"net/http"
)

func FetchArtists() ([]models.Artist, error) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch artists from API with error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}

	var artists []models.Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return artists, nil
}

func FetchLocationsByURL(url string) ([]models.Locations, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch locations from API with error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected stuats: %d", resp.StatusCode)
	}

	var concert_locations models.LocationsIndex
	if err := json.NewDecoder(resp.Body).Decode(&concert_locations); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return concert_locations.Index, nil
}

func FetchDatesByURL(url string) ([]models.Dates, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch Dates from API with error: %v", err)
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

func FetchRelationsByURL(url string) ([]models.Relation, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch Relations from API with error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Unexpected status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var relations models.RelationIndex
	if err := json.NewDecoder(resp.Body).Decode(&relations); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}
	return relations.Index, nil
}
