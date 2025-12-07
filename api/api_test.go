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
// MOCK TRANSPORTS
// ============================================================================

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

func errorTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	})
}

func statusCodeTransport(code int) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(code, ""), nil
	})
}

func invalidJSONTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(http.StatusOK, "invalid json"), nil
	})
}

// retryThenSuccessTransport fails N times per path, then succeeds
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

// ============================================================================
// TESTS - Data Validation
// ============================================================================

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
// TESTS - Endpoint Error Handling
// ============================================================================

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

/*
TestInitializeData_AllSuccess tests that InitializeData successfully loads all data (artists, locations, dates, relations) when all APIs are working.
*/
func setupInitializeDataTest() (func(), func()) {
	// setupInitializeDataTest provides two closures:
	//   reset:  explicitly clear package globals before running a test case
	//   restore: defer this to restore both the globals and the default HTTP transport after the test
	// This pattern prevents cross-test interference since tests in this file share global state
	// and mutate http.DefaultClient.Transport.
	// Save original state
	originalArtists := All_Artists
	originalLocations := All_Locations
	originalDates := All_Dates
	originalRelations := All_Relations
	// Save original transport
	originalTransport := http.DefaultClient.Transport

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
		http.DefaultClient.Transport = originalTransport
	}

	return reset, restore
}

func TestInitializeData_AllSuccess(t *testing.T) {
	reset, restore := setupInitializeDataTest()
	defer restore()
	reset()

	// Mock all endpoints as success
	restoreTransport := setMockTransport(successTransport())
	defer restoreTransport()

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

	// All endpoints fail
	restoreTransport := setMockTransport(errorTransport())
	defer restoreTransport()

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
func TestInitializeData_RetryEventuallySucceeds(t *testing.T) {
	reset, restore := setupInitializeDataTest()
	defer restore()
	reset()

	// Fail the first 2 attempts per endpoint, then succeed. InitializeData retries up to 2 times (3 total attempts)
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
	if len(All_Artists) == 0 || len(All_Locations) == 0 || len(All_Dates) == 0 || len(All_Relations) == 0 {
		t.Fatal("Expected all datasets to be loaded after retries")
	}
}

// ---- Loading status accessors ----
func TestLoadingStatusAccessors(t *testing.T) {
	SetLoadingStatus(true, false, false)
	s := GetLoadingStatus()
	if !s.IsLoading || s.IsLoaded || s.HasFailed {
		t.Errorf("unexpected status after first set: %+v", s)
	}

	SetLoadingStatus(false, true, false)
	s = GetLoadingStatus()
	if s.IsLoading || !s.IsLoaded || s.HasFailed {
		t.Errorf("unexpected status after second set: %+v", s)
	}

	SetLoadingStatus(false, false, true)
	s = GetLoadingStatus()
	if s.IsLoading || s.IsLoaded || !s.HasFailed {
		t.Errorf("unexpected status after third set: %+v", s)
	}
}

// ============================================================================
// TESTS - Loading Status
// ============================================================================

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

// ============================================================================
// TEST STATE MANAGEMENT
// ============================================================================

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
// TEST HELPERS - Mock HTTP Transport
// ============================================================================

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func setMockTransport(rt http.RoundTripper) func() {
	prev := http.DefaultClient.Transport
	http.DefaultClient.Transport = rt
	return func() { http.DefaultClient.Transport = prev }
}

func httpResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    &http.Request{},
	}
}
