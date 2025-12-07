package services

import (
	"testing"
	"time"

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

func TestGetDatesByID(t *testing.T) {
	defer setupTestData()()
	api.All_Dates = []models.Dates{{ID: 1, ConcertDates: []string{"2023-01-01"}}}

	if dates, err := GetDatesByID(1); err != nil || dates.ID != 1 {
		t.Errorf("GetDatesByID(1) failed: %v, %v", dates, err)
	}
	if _, err := GetDatesByID(999); err == nil {
		t.Error("GetDatesByID(999) should return error")
	}
}

func TestGetRelationsByID(t *testing.T) {
	defer setupTestData()()
	api.All_Relations = []models.Relations{
		{ID: 1, DatesLocations: map[string][]string{"paris-france": {"01-01-2020"}}},
	}

	rel, err := GetRelationsByID(1)
	if err != nil || rel.ID != 1 {
		t.Errorf("GetRelationsByID(1) failed: %v, %v", rel, err)
	}
	// Verify ProcessRelations was called (location formatted)
	if _, exists := rel.DatesLocations["paris-france"]; exists {
		t.Error("Location should be formatted after GetRelationsByID")
	}
	if _, err := GetRelationsByID(999); err == nil {
		t.Error("GetRelationsByID(999) should return error")
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
		want    time.Time
	}{
		{"02-01-2006", false, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"*23-08-2019", false, time.Date(2019, 8, 23, 0, 0, 0, 0, time.UTC)},
		{"02/01/2006", false, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)},
		{" 02.01.2006 ", false, time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)},
		{"", true, time.Time{}},
		{"32-01-2006", true, time.Time{}},
		{"not-a-date", true, time.Time{}},
	}

	for _, tt := range tests {
		got, err := parseDate(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseDate(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
		}
		if !tt.wantErr && !got.Equal(tt.want) {
			t.Errorf("parseDate(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestTitleCase(t *testing.T) {
	tests := map[string]string{
		"hello world": "Hello World",
		"HELLO":       "Hello",
		"":            "",
		"  spaces  ":  "Spaces",
	}
	for in, want := range tests {
		if got := titleCase(in); got != want {
			t.Errorf("titleCase(%q) = %q, want %q", in, got, want)
		}
	}
}


