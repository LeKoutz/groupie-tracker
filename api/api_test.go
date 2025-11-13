package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*
TestFetchArtists verifies that FetchArtistsWithContext succeeds and returns a non-empty list of artists.
*/
func TestFetchArtists(t *testing.T) {
	artists, err := FetchArtistsWithContext(context.Background())
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}
	if len(artists) == 0 {
		t.Errorf("Expected artists, but got empty slice")
	}
}

/*
TestFetchLocations verifies that FetchLocationsWithContext succeeds and returns a non-empty list of locations.
*/
func TestFetchLocations(t *testing.T) {
	locations, err := FetchLocationsWithContext(context.Background())
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}
	if len(locations) == 0 {
		t.Errorf("Expected locations, but got empty slice")
	}
}

/*
TestFetchDates verifies that FetchDatesWithContext succeeds and returns a non-empty list of dates.
*/
func TestFetchDates(t *testing.T) {
	dates, err := FetchDatesWithContext(context.Background())
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}
	if len(dates) == 0 {
		t.Errorf("Expected dates, but got empty slice")
	}
}

/*
TestFetchRelations verifies that FetchRelationsWithContext succeeds and returns a non-empty list of relations.
*/
func TestFetchRelations(t *testing.T) {
	relations, err := FetchRelationsWithContext(context.Background())
	if err != nil {
		t.Errorf("Did not expect an error, but got: %v", err)
	}
	if len(relations) == 0 {
		t.Errorf("Expected relations, but got empty slice")
	}
}

/*
TestFetchArtists_DataValidation checks that fetched artist data has valid fields: positive ID, non-empty name, members, creation date, first album, locations, concert dates, and relations.
*/
func TestFetchArtists_DataValidation(t *testing.T) {
	artists, err := FetchArtistsWithContext(context.Background())
	if err != nil {
		t.Skip("Skipping data validation due to fetch error")
	}
	for _, artist := range artists {
		if artist.ID <= 0 {
			t.Errorf("Artist ID should be positive, got %d", artist.ID)
		}
		if artist.Name == "" {
			t.Error("Artist Name should not be empty")
		}
		if len(artist.Members) == 0 {
			t.Error("Artist Members should not be empty")
		}
		if artist.CreationDate <= 0 {
			t.Errorf("Artist CreationDate should be positive, got %d", artist.CreationDate)
		}
		if artist.FirstAlbum == "" {
			t.Error("Artist FirstAlbum should not be empty")
		}
		if artist.Locations == "" {
			t.Error("Artist Locations should not be empty")
		}
		if artist.ConcertDates == "" {
			t.Error("Artist ConcertDates should not be empty")
		}
		if artist.Relations == "" {
			t.Error("Artist Relations should not be empty")
		}
	}
}

// Test error scenarios with mocked servers
/*
TestFetchArtists_NetworkError simulates a network error by setting an invalid URL and verifies that FetchArtistsWithContext returns an error and nil artists.
*/
func TestFetchArtists_NetworkError(t *testing.T) {
	// Save original URL and restore after test
	originalURL := ARTISTS_API
	defer func() { ARTISTS_API = originalURL }()

	// Set to invalid URL to simulate network error
	ARTISTS_API = "http://invalid.url"

	artists, err := FetchArtistsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for invalid URL, but got none")
	}
	if artists != nil {
		t.Error("Expected nil artists on error")
	}
}

/*
TestFetchArtists_StatusCodeError uses a mock server returning a 500 status code to verify that FetchArtistsWithContext returns an error and nil artists.
*/
func TestFetchArtists_StatusCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := ARTISTS_API
	ARTISTS_API = server.URL
	defer func() { ARTISTS_API = originalURL }()

	artists, err := FetchArtistsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for non-200 status code, but got none")
	}
	if artists != nil {
		t.Error("Expected nil artists on error")
	}
}

/*
TestFetchArtists_JSONDecodeError uses a mock server returning invalid JSON to verify that FetchArtistsWithContext returns a decode error and nil artists.
*/
func TestFetchArtists_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	originalURL := ARTISTS_API
	ARTISTS_API = server.URL
	defer func() { ARTISTS_API = originalURL }()

	artists, err := FetchArtistsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
	if artists != nil {
		t.Error("Expected nil artists on error")
	}
}

/*
TestFetchLocations_StatusCodeError uses a mock server returning a 404 status code to verify that FetchLocationsWithContext returns an error and nil locations.
*/
func TestFetchLocations_StatusCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	originalURL := LOCATIONS_API
	LOCATIONS_API = server.URL
	defer func() { LOCATIONS_API = originalURL }()

	locations, err := FetchLocationsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for non-200 status code, but got none")
	}
	if locations != nil {
		t.Error("Expected nil locations on error")
	}
}

/*
TestFetchLocations_JSONDecodeError uses a mock server returning invalid JSON to verify that FetchLocationsWithContext returns a decode error and nil locations.
*/
func TestFetchLocations_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	originalURL := LOCATIONS_API
	LOCATIONS_API = server.URL
	defer func() { LOCATIONS_API = originalURL }()

	locations, err := FetchLocationsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
	if locations != nil {
		t.Error("Expected nil locations on error")
	}
}

/*
TestInitializeData_AllSuccess tests that InitializeData successfully loads all data (artists, locations, dates, relations) when all APIs are working.
*/
func setupInitializeDataTest() (func(), func()) {
	// Save original state
	originalArtists := All_Artists
	originalLocations := All_Locations
	originalDates := All_Dates
	originalRelations := All_Relations
	originalArtistsURL := ARTISTS_API
	originalLocationsURL := LOCATIONS_API
	originalDatesURL := DATES_API
	originalRelationsURL := RELATIONS_API

	/*
		reset: A function you call manually to set globals to nil
		restore: A function you defer to run at the end
	*/

	reset := func() {
		All_Artists = nil
		All_Locations = nil
		All_Dates = nil
		All_Relations = nil
	}

	restore := func() {
		All_Artists = originalArtists
		All_Locations = originalLocations
		All_Dates = originalDates
		All_Relations = originalRelations
		ARTISTS_API = originalArtistsURL
		LOCATIONS_API = originalLocationsURL
		DATES_API = originalDatesURL
		RELATIONS_API = originalRelationsURL
	}

	return reset, restore
}

func TestInitializeData_AllSuccess(t *testing.T) {
	reset, restore := setupInitializeDataTest()
	defer restore()
	reset()

	// Use real APIs for this test (they should work)
	errors := InitializeData()

	if errors != nil {
		t.Errorf("Expected no errors, but got: %v", errors)
	}
	if len(All_Artists) == 0 {
		t.Error("Expected artists to be loaded")
	}
	if len(All_Locations) == 0 {
		t.Error("Expected locations to be loaded")
	}
	if len(All_Dates) == 0 {
		t.Error("Expected dates to be loaded")
	}
	if len(All_Relations) == 0 {
		t.Error("Expected relations to be loaded")
	}
}

/*
TestInitializeData_PartialFailure tests that InitializeData handles partial failures: when one API fails, it still loads data from the others and returns errors for the failed one.
*/
func TestInitializeData_PartialFailure(t *testing.T) {
	reset, restore := setupInitializeDataTest()
	defer restore()
	reset()

	// Set one API to fail
	ARTISTS_API = "http://invalid.url"

	errors := InitializeData()

	if errors == nil {
		t.Error("Expected errors due to invalid URL, but got none")
	}
	if len(errors) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(errors))
	}
	if !strings.Contains(errors[0].Error(), "FetchArtists") {
		t.Errorf("Expected error to contain 'FetchArtists', but got: %s", errors[0].Error())
	}
	// Other data should still be loaded
	if len(All_Locations) == 0 {
		t.Error("Expected locations to be loaded despite artists failure")
	}
	if len(All_Dates) == 0 {
		t.Error("Expected dates to be loaded despite artists failure")
	}
	if len(All_Relations) == 0 {
		t.Error("Expected relations to be loaded despite artists failure")
	}
}

/*
TestInitializeData_AllFailure tests that InitializeData returns errors for all failed APIs and leaves global variables empty when all APIs fail.
*/
func TestInitializeData_AllFailure(t *testing.T) {
	reset, restore := setupInitializeDataTest()
	defer restore()
	reset()

	// Set all APIs to fail
	ARTISTS_API = "http://invalid.url"
	LOCATIONS_API = "http://invalid.url"
	DATES_API = "http://invalid.url"
	RELATIONS_API = "http://invalid.url"

	errors := InitializeData()

	if errors == nil {
		t.Error("Expected errors due to all invalid URLs, but got none")
	}
	if len(errors) != 4 {
		t.Errorf("Expected 4 errors, but got %d", len(errors))
	}
	// All global variables should remain nil/empty
	if len(All_Artists) != 0 {
		t.Error("Expected artists to remain empty on all failures")
	}
	if len(All_Locations) != 0 {
		t.Error("Expected locations to remain empty on all failures")
	}
	if len(All_Dates) != 0 {
		t.Error("Expected dates to remain empty on all failures")
	}
	if len(All_Relations) != 0 {
		t.Error("Expected relations to remain empty on all failures")
	}
}

/*
TestFetchDates_StatusCodeError uses a mock server returning a 400 status code to verify that FetchDatesWithContext returns an error and nil dates.
*/
func TestFetchDates_StatusCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	originalURL := DATES_API
	DATES_API = server.URL
	defer func() { DATES_API = originalURL }()

	dates, err := FetchDatesWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for non-200 status code, but got none")
	}
	if dates != nil {
		t.Error("Expected nil dates on error")
	}
}

/*
TestFetchDates_JSONDecodeError uses a mock server returning invalid JSON to verify that FetchDatesWithContext returns a decode error and nil dates.
*/
func TestFetchDates_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	originalURL := DATES_API
	DATES_API = server.URL
	defer func() { DATES_API = originalURL }()

	dates, err := FetchDatesWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
	if dates != nil {
		t.Error("Expected nil dates on error")
	}
}

/*
TestFetchRelations_StatusCodeError uses a mock server returning a 403 status code to verify that FetchRelationsWithContext returns an error and nil relations.
*/
func TestFetchRelations_StatusCodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	originalURL := RELATIONS_API
	RELATIONS_API = server.URL
	defer func() { RELATIONS_API = originalURL }()

	relations, err := FetchRelationsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for non-200 status code, but got none")
	}
	if relations != nil {
		t.Error("Expected nil relations on error")
	}
}

/*
TestFetchRelations_JSONDecodeError uses a mock server returning invalid JSON to verify that FetchRelationsWithContext returns a decode error and nil relations.
*/
func TestFetchRelations_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	originalURL := RELATIONS_API
	RELATIONS_API = server.URL
	defer func() { RELATIONS_API = originalURL }()

	relations, err := FetchRelationsWithContext(context.Background())
	if err == nil {
		t.Error("Expected error for invalid JSON, but got none")
	}
	if relations != nil {
		t.Error("Expected nil relations on error")
	}
}

/*
TestFetchLocations_DataValidation checks that fetched location data has valid fields: positive ID, non-empty locations list, and non-empty dates string.
*/
func TestFetchLocations_DataValidation(t *testing.T) {
	locations, err := FetchLocationsWithContext(context.Background())
	if err != nil {
		t.Skip("Skipping data validation due to fetch error")
	}
	for _, loc := range locations {
		if loc.ID <= 0 {
			t.Errorf("Location ID should be positive, got %d", loc.ID)
		}
		if len(loc.Locations) == 0 {
			t.Error("Location Locations should not be empty")
		}
		if loc.Dates == "" {
			t.Error("Location Dates should not be empty")
		}
	}
}

/*
TestFetchDates_DataValidation checks that fetched date data has valid fields: positive ID and non-empty concert dates list.
*/
func TestFetchDates_DataValidation(t *testing.T) {
	dates, err := FetchDatesWithContext(context.Background())
	if err != nil {
		t.Skip("Skipping data validation due to fetch error")
	}
	for _, date := range dates {
		if date.ID <= 0 {
			t.Errorf("Date ID should be positive, got %d", date.ID)
		}
		if len(date.ConcertDates) == 0 {
			t.Error("Date ConcertDates should not be empty")
		}
	}
}

/*
TestFetchRelations_DataValidation checks that fetched relation data has valid fields: positive ID and non-empty dates locations map.
*/
func TestFetchRelations_DataValidation(t *testing.T) {
	relations, err := FetchRelationsWithContext(context.Background())
	if err != nil {
		t.Skip("Skipping data validation due to fetch error")
	}
	for _, rel := range relations {
		if rel.ID <= 0 {
			t.Errorf("Relation ID should be positive, got %d", rel.ID)
		}
		if len(rel.DatesLocations) == 0 {
			t.Error("Relation DatesLocations should not be empty")
		}
	}
}
