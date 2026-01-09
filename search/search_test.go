package search

import (
	"testing"
	"groupie-tracker/models"
)

func TestSearchAll(t *testing.T) {
	// Fake data for testing
	fakeArtists := []models.Artists{
		{
			ID:           1,
			Name:         "Queen",
			Members:      []string{"Freddie Mercury", "Brian May", "John Daecon", "Roger Meddows-Taylor", "Mike Grose", "Barry Mitchell", "Doug Fogie"},
			CreationDate: 1970,
			FirstAlbum:   "14-12-1973",
		},
	}
	// Mock the GetRelationsByID function
	fakeRelations := func(id int) (*models.Relations, error) {
		return &models.Relations{
			ID: id,
			SortedLocations: []string{"New York, USA"},
			DatesLocations: map[string][]string{
				"New York, USA":    {"2022-01-01"},
			},
		}, nil
	}

	tests := []struct {
		name           string
		query          string
		expectedResults int
		expectedLabel   string
	}{
		{
			name:           "Search by artist name",
			query:          "Queen",
			expectedResults: 1,
			expectedLabel:   "Queen - Artist/Band",
		},
		{	
			name:           "Search by member name",
			query:          "Freddie Mercury",
			expectedResults: 1,
			expectedLabel:   "Freddie Mercury - Member of Queen",
		},
		{
			name:           "Search by first album",
			query:          "14-12-1973",
			expectedResults: 1,
			expectedLabel:   "14-12-1973 - First Album of Queen",
		},
		{
			name:           "Search by creation date",
			query:          "1970",
			expectedResults: 1,
			expectedLabel:   "1970 - Creation Date of Queen",
		},
		{
			name:           "Search by location",
			query:          "New York",
			expectedResults: 1,
			expectedLabel:   "New York, USA - Concert location on 2022-01-01 for Queen",
		},
		{
			name:           "Search by date",
			query:          "2022-01-01",
			expectedResults: 1,
			expectedLabel:   "2022-01-01 - Concert date at New York, USA for Queen",
		},
		{
			name:           "No results",
			query:          "NonExistent",
			expectedResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := SearchAll(tt.query, fakeArtists, fakeRelations)
			if tt.expectedResults > 0 {
				if len(results) == 0 {
					t.Fatalf("expected at least 1 result, got 0")
				}
				if results[0].Label != tt.expectedLabel {
					t.Errorf("expected label %q, got %q", tt.expectedLabel, results[0].Label)
				}
			}
		})
	}
}
