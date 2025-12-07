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

// failOneEndpoint fails a specific endpoint, succeeds for others
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
// TESTS - InitializeData
// ============================================================================

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