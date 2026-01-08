package search

import (
	"strconv"
	"strings"

	"groupie-tracker/api"
)

type SearchResult struct {
	Label string
	ID    int
}

// SearchArtists searches artists by name, members, first album, or creation date based on the query string
func SearchArtists(query string) []SearchResult {
	results := []SearchResult{}
	searchQuery := strings.ToLower(query)
	for _, artist := range api.All_Artists {
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
	}
	return results
}
