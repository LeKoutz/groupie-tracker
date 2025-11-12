package api

import (
	"context"
	"encoding/json"
	"fmt"
	"groupie-tracker/models"
	"net/http"
	"sync"
	"time"
)

var (
	ARTISTS_API   = "https://groupietrackers.herokuapp.com/api/artists"
	LOCATIONS_API = "https://groupietrackers.herokuapp.com/api/locations"
	DATES_API     = "https://groupietrackers.herokuapp.com/api/dates"
	RELATIONS_API = "https://groupietrackers.herokuapp.com/api/relation"
)

var (
	All_Artists   []models.Artists
	All_Locations []models.Locations
	All_Dates     []models.Dates
	All_Relations []models.Relations
	Status        LoadingStatus
	statusMutex   sync.RWMutex
)

type LoadingStatus struct {
	IsLoading bool
	IsLoaded  bool
	HasFailed bool
}

// InitializeData fetches data from all APIs asynchronously.
// If a fetch fails, it retries up to 2 times.
// Each API fetch times out after 5 seconds returning an error.
func InitializeData() []error {
	var errors []error
	maxRetries := 2
	ch := make(chan error, 4)
	// Fetch Artists
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if ctx.Err() != nil {
				ch <- fmt.Errorf("FetchArtists timed out on attempt %d\n", attempt)
				return
			}
			artists, err := FetchArtistsWithContext(ctx)
			if err == nil {
				All_Artists = artists
				ch <- nil
				return
			}
			fmt.Printf("FetchArtists attempt %d failed: %v\n", attempt, err)
			if attempt < maxRetries {
				time.Sleep(1 * time.Second)
			}
		}
		ch <- fmt.Errorf("FetchArtists failed after %d attempts: %v", maxRetries+1, err)
	}()
	// Fetch Locations
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if ctx.Err() != nil {
				ch <- fmt.Errorf("FetchLocations timed out on attempt %d\n", attempt)
				return
			}
			locations, err := FetchLocationsWithContext(ctx)
			if err == nil {
				All_Locations = locations
				ch <- nil
				return
			}
			fmt.Printf("FetchLocations attempt %d failed: %v\n", attempt, err)
			if attempt < maxRetries {
				time.Sleep(1 * time.Second)
			}
		}
		ch <- fmt.Errorf("FetchLocations failed after %d attempts: %v", maxRetries+1, err)
	}()
	// Fetch Dates
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if ctx.Err() != nil {
				ch <- fmt.Errorf("FetchDates timed out on attempt %d\n", attempt)
				return
			}
			dates, err := FetchDatesWithContext(ctx)
			if err == nil {
				All_Dates = dates
				ch <- nil
				return
			}
			fmt.Printf("FetchDates attempt %d failed: %v\n", attempt, err)
			if attempt < maxRetries {
				time.Sleep(1 * time.Second)
			}
		}
		ch <- fmt.Errorf("FetchDates failed after %d attempts: %v", maxRetries+1, err)
	}()
	// Fetch Relations
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if ctx.Err() != nil {
				ch <- fmt.Errorf("FetchRelations timed out on attempt %d\n", attempt)
				return
			}
			relations, err := FetchRelationsWithContext(ctx)
			if err == nil {
				All_Relations = relations
				ch <- nil
				return
			}
			fmt.Printf("FetchRelations attempt %d failed: %v\n", attempt, err)
			if attempt < maxRetries {
				time.Sleep(1 * time.Second)
			}
		}
		ch <- fmt.Errorf("FetchRelations failed after %d attempts: %v", maxRetries+1, err)
	}()
	// Collect results
	for i := 0; i < 4; i++ {
		if err := <-ch; err != nil {
			errors = append(errors, err)
		}
	}
	close(ch)
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func FetchArtistsWithContext(ctx context.Context) ([]models.Artists, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ARTISTS_API, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func FetchLocationsWithContext(ctx context.Context) ([]models.Locations, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LOCATIONS_API, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func FetchDatesWithContext(ctx context.Context) ([]models.Dates, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, DATES_API, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func FetchRelationsWithContext(ctx context.Context) ([]models.Relations, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, RELATIONS_API, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
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

func SetLoadingStatus(loading, loaded, failed bool) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	Status = LoadingStatus{
		IsLoading: loading,
		IsLoaded:  loaded,
		HasFailed: failed,
	}
}

func GetLoadingStatus() LoadingStatus {
	statusMutex.RLock()
	defer statusMutex.RUnlock()
	return Status
}

// RefreshData automatically refreshes the API data every 24hours if the fetch succeeded,
// or every 1 second if it failed
func RefreshData() {
	GetLoadingStatus()
	for {
		if GetLoadingStatus().IsLoading {
			continue
		} else if GetLoadingStatus().IsLoaded {
			time.Sleep(24 * time.Hour)
			fmt.Println("Refreshing data...")
			SetLoadingStatus(true, false, false)
			err := InitializeData()
			if err != nil {
				SetLoadingStatus(false, false, true)
				continue
			} else {
				SetLoadingStatus(false, true, false)
				continue
			}
		} else if GetLoadingStatus().HasFailed {
			time.Sleep(1 * time.Second)
			fmt.Println("Refreshing data...")
			SetLoadingStatus(true, false, false)
			err := InitializeData()
			if err != nil {
				SetLoadingStatus(false, false, true)
				continue
			} else {
				SetLoadingStatus(false, true, false)
				continue
			}
		}
	}
}
