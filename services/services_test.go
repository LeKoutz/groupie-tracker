package services

import (
	"testing"

	"groupie-tracker/api"
	"groupie-tracker/models"
)

func setupTestData() func() {
	origArtists, origLocations := api.All_Artists, api.All_Locations
	origDates, origRelations := api.All_Dates, api.All_Relations
	return func() {
		api.All_Artists, api.All_Locations = origArtists, origLocations
		api.All_Dates, api.All_Relations = origDates, origRelations
	}
}

func TestGetArtistByID(t *testing.T) {
	defer setupTestData()()
	api.All_Artists = []models.Artists{
		{ID: 1, Name: "Artist 1"},
		{ID: 2, Name: "Artist 2"},
	}

	if artist, err := GetArtistByID(1); err != nil || artist.ID != 1 {
		t.Errorf("GetArtistByID(1) = %v, %v; want artist with ID 1, nil", artist, err)
	}

	if _, err := GetArtistByID(999); err == nil {
		t.Error("GetArtistByID(999) should return error")
	}
}

func TestGetLocationsByID(t *testing.T) {
	defer setupTestData()()
	api.All_Locations = []models.Locations{{ID: 1, Locations: []string{"Loc1"}}}

	if loc, err := GetLocationsByID(1); err != nil || loc.ID != 1 {
		t.Errorf("GetLocationsByID(1) failed: %v, %v", loc, err)
	}
	if _, err := GetLocationsByID(999); err == nil {
		t.Error("GetLocationsByID(999) should return error")
	}
}
