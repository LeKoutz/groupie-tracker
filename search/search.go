package search

import (
	"strconv"
	"strings"

	"groupie-tracker/models"
)

type SearchResult struct {
	Label string
	ID    int
}

// SearchAll searches artists by name, members, first album, creation date, locations, and dates based on the query string
func SearchAll(query string, artists []models.Artists, getRelations func(int) (*models.Relations, error)) []SearchResult {
	results := []SearchResult{}
	searchQuery := strings.ToLower(query)
	for _, artist := range artists {
		// Search by name
		if strings.Contains(strings.ToLower(artist.Name), searchQuery) {
			results = append(results, SearchResult{
				Label: artist.Name + " - Artist/Band",
				ID:    artist.ID,
			})
		}
		// Search by members
		for _, member := range artist.Members {
			if strings.Contains(strings.ToLower(member), searchQuery) {
				results = append(results, SearchResult{
					Label: member + " - Member of " + artist.Name,
					ID:    artist.ID,
				})
			}
		}
		// Search by first album
		if strings.Contains(strings.ToLower(artist.FirstAlbum), searchQuery) {
			results = append(results, SearchResult{
				Label: artist.FirstAlbum + " - First Album of " + artist.Name,
				ID:    artist.ID,
			})
		}
		// Search by creation date
		creationDateStr := strconv.Itoa(artist.CreationDate)
		if strings.Contains(creationDateStr, searchQuery) {
			results = append(results, SearchResult{
				Label: creationDateStr + " - Creation Date of " + artist.Name,
				ID:    artist.ID,
			})
		}
		rel, err := getRelations(artist.ID)
		if err != nil {
			continue
		}
		// Search by locations in relations
		for _, loc := range rel.SortedLocations {
			dates := rel.DatesLocations[loc]
			// Search by dates in relations
			for _, date := range dates {
				if strings.Contains(strings.ToLower(date), searchQuery) {
					results = append(results, SearchResult{
						Label: date + " - Concert date at " + loc + " for " + artist.Name,
						ID:    artist.ID,
					})
				}
				if strings.Contains(strings.ToLower(loc), searchQuery) || strings.Contains(normalize(loc), normalize(query)) {
					results = append(results, SearchResult{
						Label: loc + " - Concert location on " + date + " for " + artist.Name,
						ID:    artist.ID,
					})
				}
			}
		}
	}
	return results
}
// removes punctuation and spaces from string
func normalize(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}