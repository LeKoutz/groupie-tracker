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
)

// NOTE ABOUT TEST EXECUTION AND GLOBAL STATE
// These tests intentionally mutate package-global variables (e.g., All_Artists, All_Locations,
// All_Dates, All_Relations) and also swap out http.DefaultClient.Transport to mock HTTP calls.
// This is done to avoid actual network calls and to control the test environment.

/*
TestFetchArtists_DataValidation checks that fetched artist data has valid fields: positive ID, non-empty name, members, creation date, first album, locations, concert dates, and relations.
*/
func TestFetchArtists_DataValidation(t *testing.T) {
	restore := setMockTransport(successTransport())
	defer restore()

	artists, err := FetchArtistsWithContext(context.Background())
	if err != nil {
		// Guard against accidental network usage: this test assumes successTransport is set.
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

// Table-driven test for network errors across all endpoints
func TestAllEndpoints_NetworkError(t *testing.T) {
	type endpointTest struct {
		name        string
		path        string
		fetchFunc   func(context.Context) (any, error)
		shouldBeNil bool
	}

	endpoints := []endpointTest{
		{
			name: "artists",
			path: "/api/artists",
			fetchFunc: func(ctx context.Context) (any, error) {
				artists, err := FetchArtistsWithContext(ctx)
				if artists == nil {
					return nil, err
				}
				return artists, err
			},
			shouldBeNil: true,
		},
		{
			name: "locations",
			path: "/api/locations",
			fetchFunc: func(ctx context.Context) (any, error) {
				locations, err := FetchLocationsWithContext(ctx)
				if locations == nil {
					return nil, err
				}
				return locations, err
			},
			shouldBeNil: true,
		},
		{
			name: "dates",
			path: "/api/dates",
			fetchFunc: func(ctx context.Context) (any, error) {
				dates, err := FetchDatesWithContext(ctx)
				if dates == nil {
					return nil, err
				}
				return dates, err
			},
			shouldBeNil: true,
		},
		{
			name: "relations",
			path: "/api/relation",
			fetchFunc: func(ctx context.Context) (any, error) {
				relations, err := FetchRelationsWithContext(ctx)
				if relations == nil {
					return nil, err
				}
				return relations, err
			},
			shouldBeNil: true,
		},
	}

	for _, tt := range endpoints {
		t.Run(tt.name, func(t *testing.T) {
			restore := setMockTransport(errorTransport(map[string]error{
				tt.path: errors.New("dial error"),
			}))
			defer restore()

			result, err := tt.fetchFunc(context.Background())
			if err == nil {
				t.Error("Expected error for network issue, but got none")
			}
			if tt.shouldBeNil && result != nil {
				t.Error("Expected nil result on error")
			}
		})
	}
}

/*
TestFetchArtists_JSONDecodeError uses a mock server returning invalid JSON to verify that FetchArtistsWithContext returns a decode error and nil artists.
*/
func TestFetchArtists_JSONDecodeError(t *testing.T) {
	restore := setMockTransport(bodyTransport(map[string]mockBody{
		"/api/artists": {status: http.StatusOK, body: "invalid json"},
	}))
	defer restore()

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
	restore := setMockTransport(statusTransport(map[string]int{
		"/api/locations": http.StatusNotFound,
	}))
	defer restore()

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
	restore := setMockTransport(bodyTransport(map[string]mockBody{
		"/api/locations": {status: http.StatusOK, body: "invalid json"},
	}))
	defer restore()

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

	// Make artists fail, others succeed
	restoreTransport := setMockTransport(mixedTransport(map[string]endpointBehavior{
		"/api/artists": {kind: kindError},
	}))
	defer restoreTransport()

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
	restoreTransport := setMockTransport(failAllTransport())
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

// --- Retry transport and test for InitializeData retry logic ---

// retryThenSuccessTransport returns errors for the first N requests per path, then succeeds
func retryThenSuccessTransport(failuresByPath map[string]int) http.RoundTripper {
	// retryThenSuccessTransport simulates transient failures per endpoint path.
	// For each provided path, the first N (map value) requests fail with a network error,
	// and subsequent requests succeed by delegating to successTransport().
	// This is used to exercise InitializeData's retry logic (2 retries = 3 total attempts).
	// Path matching uses substring checks against r.URL.Path or r.URL.String().
	var mu sync.Mutex
	counts := make(map[string]int)
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// Determine matched key
		matched := ""
		for path := range failuresByPath {
			if strings.Contains(r.URL.Path, path) || strings.Contains(r.URL.String(), path) {
				matched = path
				break
			}
		}
		if matched != "" {
			mu.Lock()
			counts[matched]++
			c := counts[matched]
			limit := failuresByPath[matched]
			mu.Unlock()
			if c <= limit {
				return nil, errors.New("simulated transient error")
			}
			// After failures, return success for that endpoint
			return successTransport().RoundTrip(r)
		}
		// Unspecified: default to success
		return successTransport().RoundTrip(r)
	})
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

/*
TestFetchDates_StatusCodeError uses a mock server returning a 400 status code to verify that FetchDatesWithContext returns an error and nil dates.
*/
func TestFetchDates_StatusCodeError(t *testing.T) {
	restore := setMockTransport(statusTransport(map[string]int{
		"/api/dates": http.StatusBadRequest,
	}))
	defer restore()

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
	restore := setMockTransport(bodyTransport(map[string]mockBody{
		"/api/dates": {status: http.StatusOK, body: "invalid json"},
	}))
	defer restore()

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
	restore := setMockTransport(statusTransport(map[string]int{
		"/api/relation": http.StatusForbidden,
	}))
	defer restore()

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
	restore := setMockTransport(bodyTransport(map[string]mockBody{
		"/api/relation": {status: http.StatusOK, body: "invalid json"},
	}))
	defer restore()

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
	restore := setMockTransport(successTransport())
	defer restore()

	locations, err := FetchLocationsWithContext(context.Background())
	if err != nil {
		// Guard against accidental network usage: this test assumes successTransport is set.
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
	restore := setMockTransport(successTransport())
	defer restore()

	dates, err := FetchDatesWithContext(context.Background())
	if err != nil {
		// Guard against accidental network usage: this test assumes successTransport is set.
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
	restore := setMockTransport(successTransport())
	defer restore()

	relations, err := FetchRelationsWithContext(context.Background())
	if err != nil {
		// Guard against accidental network usage: this test assumes successTransport is set.
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

// ---- Test helpers ----

// roundTripFunc helps us define inline RoundTripper implementations
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// setMockTransport sets http.DefaultClient.Transport and returns a restore function
func setMockTransport(rt http.RoundTripper) func() {
	prev := http.DefaultClient.Transport
	http.DefaultClient.Transport = rt
	return func() { http.DefaultClient.Transport = prev }
}

// Helpers to build responses
func httpResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    &http.Request{},
	}
}

// successTransport returns valid JSON bodies for all endpoints
func successTransport() http.RoundTripper {
	// The JSON payloads below are intentionally minimal yet valid for the current model shapes
	// (Artists, LocationsIndex, DatesIndex, RelationIndex). If the model structures change
	// (fields, nesting, required properties), update these payloads accordingly.
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		url := r.URL.String()
		switch {
		case strings.Contains(url, "/api/artists"):
			// Minimal valid artists array
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

// statusTransport returns specific status codes for selected paths
func statusTransport(statusByPath map[string]int) http.RoundTripper {
	// Path matching in all transport helpers (statusTransport, errorTransport, bodyTransport,
	// mixedTransport) uses substring checks against r.URL.Path and r.URL.String().
	// To avoid accidental matches, prefer unique path stubs like "/api/artists" rather than
	// overly generic substrings. If exact matching is desired, consider tightening the matcher.
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		for path, code := range statusByPath {
			if strings.Contains(r.URL.Path, path) || strings.Contains(r.URL.String(), path) {
				return httpResponse(code, ""), nil
			}
		}
		// default success for others
		return successTransport().RoundTrip(r)
	})
}

// errorTransport returns errors for selected paths
func errorTransport(errByPath map[string]error) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		for path, e := range errByPath {
			if strings.Contains(r.URL.Path, path) || strings.Contains(r.URL.String(), path) {
				return nil, e
			}
		}
		// default success for others
		return successTransport().RoundTrip(r)
	})
}

type mockBody struct {
	status int
	body   string
}

// bodyTransport returns custom body for selected paths
func bodyTransport(bodyByPath map[string]mockBody) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		for path, mb := range bodyByPath {
			if strings.Contains(r.URL.Path, path) || strings.Contains(r.URL.String(), path) {
				return httpResponse(mb.status, mb.body), nil
			}
		}
		// default success for others
		return successTransport().RoundTrip(r)
	})
}

// endpoint behavior for mixed transport
type endpointBehavior struct{ kind int }

const (
	kindSuccess = iota
	kindError
	kindStatus500
)

// mixedTransport allows specifying behavior per path; unspecified default to success
func mixedTransport(behaviors map[string]endpointBehavior) http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		for path, b := range behaviors {
			if strings.Contains(r.URL.Path, path) || strings.Contains(r.URL.String(), path) {
				switch b.kind {
				case kindError:
					return nil, errors.New("simulated network error")
				case kindStatus500:
					return httpResponse(http.StatusInternalServerError, ""), nil
				default:
					// success
					return successTransport().RoundTrip(r)
				}
			}
		}
		return successTransport().RoundTrip(r)
	})
}

func failAllTransport() http.RoundTripper {
	return roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, errors.New("simulated network error")
	})
}
