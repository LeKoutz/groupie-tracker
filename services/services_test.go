package services

import (
	"reflect"
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

func TestFormatLocationName(t *testing.T) {
	tests := map[string]string{
		"new-york-usa":         "New York, USA",
		"san-juan-puerto-rico": "San Juan Puerto, Rico",
		"los_angeles-usa":      "Los Angeles, USA",
		"london":               "London",
		"london-uk":            "London, UK",
	}
	for in, want := range tests {
		if got := formatLocationName(in); got != want {
			t.Errorf("formatLocationName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDateNewer(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"01-01-2021", "01-01-2020", true},
		{"01-01-2020", "01-01-2021", false},
		{"invalid", "01-01-2020", false},
		{"01-01-2020", "invalid", true},
	}
	for _, tt := range tests {
		if got := dateNewer(tt.a, tt.b); got != tt.want {
			t.Errorf("dateNewer(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestSortDatesInLocations(t *testing.T) {
	r := &models.Relations{
		DatesLocations: map[string][]string{
			"loc1": {"01-01-2020", "02-01-2021", "bad-date", "15-02-2019"},
			"loc2": {"bad"},
			"loc3": {"03-03-2022", "01-01-2020"},
		},
	}

	sortDatesInLocations(r)

	want1 := []string{"02-01-2021", "01-01-2020", "15-02-2019", "bad-date"}
	if !reflect.DeepEqual(r.DatesLocations["loc1"], want1) {
		t.Errorf("loc1 = %v, want %v", r.DatesLocations["loc1"], want1)
	}

	want3 := []string{"03-03-2022", "01-01-2020"}
	if !reflect.DeepEqual(r.DatesLocations["loc3"], want3) {
		t.Errorf("loc3 = %v, want %v", r.DatesLocations["loc3"], want3)
	}
}

func TestSortLocationsByDate(t *testing.T) {
	r := &models.Relations{
		DatesLocations: map[string][]string{
			"Location A": {"01-01-2020"},
			"Location B": {"15-06-2021"},
			"Location C": {},
		},
	}

	sortLocationsByDate(r)

	if len(r.SortedLocations) != 2 {
		t.Errorf("Expected 2 locations (C has no dates), got %d", len(r.SortedLocations))
	}
	if r.SortedLocations[0] != "Location B" {
		t.Errorf("Expected Location B first (newest), got %s", r.SortedLocations[0])
	}
}

func TestFormatLocations(t *testing.T) {
	r := &models.Relations{
		DatesLocations: map[string][]string{
			"new-york-usa": {"01-01-2020"},
			"paris-france": {"02-01-2020"},
		},
	}

	formatLocations(r)

	if _, exists := r.DatesLocations["new-york-usa"]; exists {
		t.Error("Raw key should be replaced with formatted key")
	}
	if _, exists := r.DatesLocations["New York, USA"]; !exists {
		t.Error("Formatted key should exist")
	}
}

func TestFormatLocationNameAlreadyFormatted(t *testing.T) {
    // Test that already formatted strings are not processed again
    formatted := "New York, USA"
    result := formatLocationName(formatted)
    if result != formatted {
        t.Errorf("Already formatted string should be returned as-is, got %q", result)
    }
}


func TestProcessRelations(t *testing.T) {
	r := &models.Relations{
		DatesLocations: map[string][]string{
			"new-york-usa": {"01-01-2020", "15-06-2021"},
			"paris-france": {"03-03-2022", "01-01-2019"},
		},
	}

	ProcessRelations(r)

	// Check formatting
	if _, exists := r.DatesLocations["New York, USA"]; !exists {
		t.Error("Locations should be formatted")
	}

	// Check date sorting (newest first)
	nydates := r.DatesLocations["New York, USA"]
	if nydates[0] != "15-06-2021" {
		t.Errorf("Dates not sorted, got %v", nydates)
	}

	// Check location sorting (Paris has newest date: 2022)
	if len(r.SortedLocations) == 0 {
		t.Error("SortedLocations should not be empty")
	} else if r.SortedLocations[0] != "Paris, France" {
		t.Errorf("Expected Paris first, got %s", r.SortedLocations[0])
	}
}
