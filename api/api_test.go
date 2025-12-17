package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// TEST HELPERS - Mock HTTP Transport
// ============================================================================
//
// This section contains helper functions for mocking HTTP requests during testing.
// These functions allow us to test API behavior without making actual network calls.

type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper interface for our mock transport.
// It delegates to the provided function to generate responses.
func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

// setMockTransport replaces the default HTTP transport with a mock one.
// Returns a restore function to revert the transport back to its original state.
// This ensures tests are isolated and don't affect each other.
func setMockTransport(rt http.RoundTripper) func() {
	prev := http.DefaultClient.Transport
	http.DefaultClient.Transport = rt
	return func() { http.DefaultClient.Transport = prev }
}

// httpResponse creates a mock HTTP response with the given status code and body.
// Used by mock transports to simulate API responses.
func httpResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    &http.Request{},
	}
}

// ============================================================================
// MOCK TRANSPORTS
// ============================================================================
//
// This section contains mock HTTP transports that simulate different API scenarios.
// Each transport returns predefined responses for testing various conditions.

// successTransport simulates successful API responses for all endpoints.
// Returns valid JSON data matching the expected API response format.
func successTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		url := r.URL.String()
		switch {
		case strings.Contains(url, "/api/artists"):
			body := `[{"id":1,"image":"img","name":"Band","members":["A","B"],"creationDate":2000,"firstAlbum":"01-01-2001","locations":"/api/locations/1","concertDates":"/api/dates/1","relations":"/api/relation/1"}]`
			return httpResponse(http.StatusOK, body), nil
		case strings.Contains(url, "/api/locations"):
			body := `{"index":[{"id":1,"locations":["paris-france"],"dates":"/api/dates/1"}]}`
			return httpResponse(http.StatusOK, body), nil
		case strings.Contains(url, "/api/dates"):
			body := `{"index":[{"id":1,"dates":["01-01-2020"]}]}`
			return httpResponse(http.StatusOK, body), nil
		case strings.Contains(url, "/api/relation"):
			body := `{"index":[{"id":1,"datesLocations":{"paris-france":["01-01-2020"]}}]}`
			return httpResponse(http.StatusOK, body), nil
		default:
			return httpResponse(http.StatusNotFound, ""), nil
		}
	})
}

// errorTransport simulates network-level errors.
// Returns an error without attempting to create a response.
func errorTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	})
}

// statusCodeTransport simulates API responses with specific HTTP status codes.
// Useful for testing error handling for different status codes.
func statusCodeTransport(code int) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(code, ""), nil
	})
}

// invalidJSONTransport simulates APIs returning malformed JSON.
// Returns a 200 OK status with invalid JSON content.
func invalidJSONTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(http.StatusOK, "invalid json"), nil
	})
}

// retryThenSuccessTransport simulates transient errors that eventually succeed.
// Fails N times for each specified path, then succeeds on subsequent requests.
// This is used to test the retry mechanism in InitializeData.
func retryThenSuccessTransport(failuresPerPath map[string]int) http.RoundTripper {
	var mu sync.Mutex
	counts := make(map[string]int)
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		matched := ""
		for path := range failuresPerPath {
			if strings.Contains(r.URL.String(), path) {
				matched = path
				break
			}
		}
		if matched != "" {
			mu.Lock()
			counts[matched]++
			c := counts[matched]
			limit := failuresPerPath[matched]
			mu.Unlock()
			if c <= limit {
				return nil, errors.New("simulated transient error")
			}
		}
		return successTransport().RoundTrip(r)
	})
}

// failOneEndpoint simulates failure for a specific endpoint while others succeed.
// Useful for testing partial failure scenarios in InitializeData.
func failOneEndpoint(path string) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if strings.Contains(r.URL.String(), path) {
			return nil, errors.New("simulated error")
		}
		return successTransport().RoundTrip(r)
	})
}

// ============================================================================
// TEST STATE MANAGEMENT
// ============================================================================
//
// This section contains functions for managing test state.
// It ensures tests are isolated and don't interfere with each other.

// setupTest provides functions to reset and restore test state.
// Returns:
//   - reset: Clears all global data structures (All_Artists, All_Locations, etc.)
//   - restore: Restores the original state after tests complete
//
// This ensures each test starts with a clean slate and doesn't affect other tests.
func setupTest() (reset func(), restore func()) {
	originalArtists := All_Artists
	originalLocations := All_Locations
	originalDates := All_Dates
	originalRelations := All_Relations
	originalTransport := http.DefaultClient.Transport

	reset = func() {
		All_Artists = nil
		All_Locations = nil
		All_Dates = nil
		All_Relations = nil
	}

	restore = func() {
		All_Artists = originalArtists
		All_Locations = originalLocations
		All_Dates = originalDates
		All_Relations = originalRelations
		http.DefaultClient.Transport = originalTransport
	}

	return reset, restore
}

// ============================================================================
// TESTS - Endpoint Error Handling
// ============================================================================
//
// This section tests error handling for individual API endpoints.
// Each test verifies that the fetch functions properly handle:
//   - Network errors
//   - Invalid HTTP status codes
//   - Malformed JSON responses
//   - Successful responses

// TestFetchArtists_Errors tests FetchArtistsWithContext with various error conditions.
// Tests network errors, bad status codes, invalid JSON, and successful responses.
func TestFetchArtists_Errors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		wantErr   bool
	}{
		{"network error", errorTransport(), true},
		{"bad status code", statusCodeTransport(http.StatusInternalServerError), true},
		{"invalid JSON", invalidJSONTransport(), true},
		{"success", successTransport(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := setMockTransport(tt.transport)
			defer restore()

			result, err := FetchArtistsWithContext(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && len(result) == 0 {
				t.Error("Expected non-empty result on success")
			}
		})
	}
}

func TestFetchLocations_Errors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		wantErr   bool
	}{
		{"network error", errorTransport(), true},
		{"bad status code", statusCodeTransport(http.StatusNotFound), true},
		{"invalid JSON", invalidJSONTransport(), true},
		{"success", successTransport(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := setMockTransport(tt.transport)
			defer restore()

			result, err := FetchLocationsWithContext(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && len(result) == 0 {
				t.Error("Expected non-empty result on success")
			}
		})
	}
}

func TestFetchDates_Errors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		wantErr   bool
	}{
		{"network error", errorTransport(), true},
		{"bad status code", statusCodeTransport(http.StatusBadRequest), true},
		{"invalid JSON", invalidJSONTransport(), true},
		{"success", successTransport(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := setMockTransport(tt.transport)
			defer restore()

			result, err := FetchDatesWithContext(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && len(result) == 0 {
				t.Error("Expected non-empty result on success")
			}
		})
	}
}

func TestFetchRelations_Errors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		wantErr   bool
	}{
		{"network error", errorTransport(), true},
		{"bad status code", statusCodeTransport(http.StatusForbidden), true},
		{"invalid JSON", invalidJSONTransport(), true},
		{"success", successTransport(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := setMockTransport(tt.transport)
			defer restore()

			result, err := FetchRelationsWithContext(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && len(result) == 0 {
				t.Error("Expected non-empty result on success")
			}
		})
	}
}

// ============================================================================
// TESTS - Data Validation
// ============================================================================
//
// This section tests data validation for successfully fetched API responses.
// Each test verifies that the data structure and values are valid.

// TestFetchArtists_DataValidation verifies that fetched artist data is valid.
// Checks that each artist has:
//   - Positive ID
//   - Non-empty name
//   - Non-empty members list
//   - Valid creation date
func TestFetchArtists_DataValidation(t *testing.T) {
	restore := setMockTransport(successTransport())
	defer restore()

	artists, err := FetchArtistsWithContext(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	for _, artist := range artists {
		if artist.ID <= 0 {
			t.Errorf("Invalid artist ID: %d", artist.ID)
		}
		if artist.Name == "" {
			t.Error("Artist name is empty")
		}
		if len(artist.Members) == 0 {
			t.Error("Artist has no members")
		}
		if artist.CreationDate <= 0 {
			t.Errorf("Invalid creation date: %d", artist.CreationDate)
		}
	}
}

// TestFetchLocations_DataValidation verifies that fetched location data is valid.
// Checks that each location has:
//   - Positive ID
//   - Non-empty locations list
func TestFetchLocations_DataValidation(t *testing.T) {
	restore := setMockTransport(successTransport())
	defer restore()

	locations, err := FetchLocationsWithContext(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	for _, loc := range locations {
		if loc.ID <= 0 {
			t.Errorf("Invalid location ID: %d", loc.ID)
		}
		if len(loc.Locations) == 0 {
			t.Error("Location has no locations")
		}
	}
}

// TestFetchDates_DataValidation verifies that fetched date data is valid.
// Checks that each date has:
//   - Positive ID
//   - Non-empty concert dates list
func TestFetchDates_DataValidation(t *testing.T) {
	restore := setMockTransport(successTransport())
	defer restore()

	dates, err := FetchDatesWithContext(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	for _, date := range dates {
		if date.ID <= 0 {
			t.Errorf("Invalid date ID: %d", date.ID)
		}
		if len(date.ConcertDates) == 0 {
			t.Error("Date has no concert dates")
		}
	}
}

// TestFetchRelations_DataValidation verifies that fetched relation data is valid.
// Checks that each relation has:
//   - Positive ID
//   - Non-empty dates-locations mapping
func TestFetchRelations_DataValidation(t *testing.T) {
	restore := setMockTransport(successTransport())
	defer restore()

	relations, err := FetchRelationsWithContext(context.Background())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	for _, rel := range relations {
		if rel.ID <= 0 {
			t.Errorf("Invalid relation ID: %d", rel.ID)
		}
		if len(rel.DatesLocations) == 0 {
			t.Error("Relation has no dates-locations mapping")
		}
	}
}

// ============================================================================
// TESTS - InitializeData
// ============================================================================
//
// This section tests the InitializeData function which fetches all API data.
// Tests cover:
//   - All endpoints succeeding
//   - Partial failures (some endpoints fail)
//   - Complete failures (all endpoints fail)
//   - Retry mechanism working correctly

// TestInitializeData_AllSuccess tests that InitializeData successfully loads
// data from all endpoints when all requests succeed.
func TestInitializeData_AllSuccess(t *testing.T) {
	reset, restore := setupTest()
	defer restore()
	reset()

	restoreTransport := setMockTransport(successTransport())
	defer restoreTransport()

	errs := InitializeData()

	if errs != nil {
		t.Errorf("Expected no errors, got: %v", errs)
	}
	if len(All_Artists) == 0 || len(All_Locations) == 0 ||
		len(All_Dates) == 0 || len(All_Relations) == 0 {
		t.Error("Expected all data to be loaded")
	}
}

// TestInitializeData_PartialFailure tests that InitializeData handles
// partial failures gracefully. When one endpoint fails, others should still load.
func TestInitializeData_PartialFailure(t *testing.T) {
	reset, restore := setupTest()
	defer restore()
	reset()

	restoreTransport := setMockTransport(failOneEndpoint("/api/artists"))
	defer restoreTransport()

	errs := InitializeData()

	if errs == nil || len(errs) != 1 {
		t.Errorf("Expected 1 error, got: %v", errs)
	}
	if errs != nil && !strings.Contains(errs[0].Error(), "FetchArtists") {
		t.Errorf("Expected FetchArtists error, got: %v", errs[0])
	}
	// Other data should still load
	if len(All_Locations) == 0 || len(All_Dates) == 0 || len(All_Relations) == 0 {
		t.Error("Expected other data to load despite artist failure")
	}
}

// TestInitializeData_AllFailure tests that InitializeData properly
// collects all errors when all endpoints fail.
func TestInitializeData_AllFailure(t *testing.T) {
	reset, restore := setupTest()
	defer restore()
	reset()

	restoreTransport := setMockTransport(errorTransport())
	defer restoreTransport()

	errs := InitializeData()

	if errs == nil || len(errs) != 4 {
		t.Errorf("Expected 4 errors, got: %v", errs)
	}
	if len(All_Artists) != 0 || len(All_Locations) != 0 ||
		len(All_Dates) != 0 || len(All_Relations) != 0 {
		t.Error("Expected all data to remain empty on failure")
	}
}

// TestInitializeData_RetrySuccess tests that InitializeData properly
// retries failed requests and eventually succeeds.
// Each endpoint is configured to fail 2 times, then succeed on the 3rd attempt.
func TestInitializeData_RetrySuccess(t *testing.T) {
	reset, restore := setupTest()
	defer restore()
	reset()

	// Fail twice, succeed on third attempt (maxRetries=2 means 3 total attempts)
	rt := retryThenSuccessTransport(map[string]int{
		"/api/artists":   2,
		"/api/locations": 2,
		"/api/dates":     2,
		"/api/relation":  2,
	})
	restoreTransport := setMockTransport(rt)
	defer restoreTransport()

	errs := InitializeData()

	if errs != nil {
		t.Fatalf("Expected no errors after retries, got: %v", errs)
	}
	if len(All_Artists) == 0 || len(All_Locations) == 0 ||
		len(All_Dates) == 0 || len(All_Relations) == 0 {
		t.Error("Expected all data to be loaded after retries")
	}
}

// ============================================================================
// TESTS - Loading Status
// ============================================================================
//
// This section tests the loading status management functions.
// Verifies that SetLoadingStatus and GetLoadingStatus work correctly
// and that the status is properly tracked.

// TestLoadingStatus verifies that loading status is correctly set and retrieved.
// Tests all three states: loading, loaded, and failed.
func TestLoadingStatus(t *testing.T) {
	SetLoadingStatus(true, false, false)
	s := GetLoadingStatus()
	if !s.IsLoading || s.IsLoaded || s.HasFailed {
		t.Errorf("Unexpected status: %+v", s)
	}

	SetLoadingStatus(false, true, false)
	s = GetLoadingStatus()
	if s.IsLoading || !s.IsLoaded || s.HasFailed {
		t.Errorf("Unexpected status: %+v", s)
	}

	SetLoadingStatus(false, false, true)
	s = GetLoadingStatus()
	if s.IsLoading || s.IsLoaded || !s.HasFailed {
		t.Errorf("Unexpected status: %+v", s)
	}
}

// ============================================================================
// TESTS - RefreshData
// ============================================================================
//
// This section tests the RefreshData function which automatically refreshes
// API data at intervals. Tests verify:
//   - No refresh when already loading
//   - Retry behavior after failure

// TestRefreshData_NoRefreshWhenLoading verifies that RefreshData doesn't
// start a new refresh if one is already in progress.
// The function should wait until the current loading completes.
func TestRefreshData_NoRefreshWhenLoading(t *testing.T) {
	SetLoadingStatus(true, false, false)

	refreshStarted := make(chan bool, 1)
	go func() {
		RefreshData()
		refreshStarted <- true
	}()

	select {
	case <-refreshStarted:
		t.Error("RefreshData should not start when already loading")
	case <-time.After(100 * time.Millisecond):
		// Expected: RefreshData should be waiting
	}

	SetLoadingStatus(false, false, false)
}

// TestRefreshData_RetryOnFailure verifies that RefreshData properly
// retries fetching data after a failure.
// When data fetch fails, it should continue trying to refresh.
func TestRefreshData_RetryOnFailure(t *testing.T) {
	reset, restore := setupTest()
	defer restore()
	reset()

	SetLoadingStatus(false, false, true)
	http.DefaultClient.Transport = errorTransport()

	go RefreshData()

	// Wait for retry to happen (1 second sleep + processing)
	time.Sleep(1500 * time.Millisecond)

	status := GetLoadingStatus()
	if !status.IsLoading {
		t.Error("Expected loading state when retrying after failure")
	}
}