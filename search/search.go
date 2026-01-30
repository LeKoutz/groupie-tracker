package search

import (
	"strconv"
	"strings"

	"groupie-tracker/models"
)

type SearchResult struct {
	Label string
	ID    int
	Category string
}

// SearchAll searches artists by name, members, first album, creation date, locations, and dates based on the query string.
// It expects a single-word query. Thus, queries like "Freddie Mercury" should be split into []string{"Freddie" "Mercury"}
func SearchAll(query string, artists []models.Artists, getRelations func(int) (*models.Relations, error)) []SearchResult {
	results := []SearchResult{}
	searchQuery := strings.ToLower(query)
	for _, artist := range artists {
		// Search by name
		for _, part := range strings.Fields(strings.ToLower(artist.Name)) {
			if strings.HasPrefix(part, searchQuery) {
				results = append(results, SearchResult{
					Label:    artist.Name + " - Artist/Band",
					ID:       artist.ID,
					Category: "artist",
				})
			}
		}
		// Search by members
		for _, member := range artist.Members {
			for _, part := range strings.Fields(strings.ToLower(member)) {
				if strings.HasPrefix(part, searchQuery) {
					results = append(results, SearchResult{
						Label:    member + " - Member of " + artist.Name,
						ID:       artist.ID,
						Category: "member",
					})
				}
			}
		}
		// Search by first album
		if strings.HasPrefix(artist.FirstAlbum, searchQuery) {
			results = append(results, SearchResult{
				Label:    artist.FirstAlbum + " - First Album of " + artist.Name,
				ID:       artist.ID,
				Category: "first_album",
			})
		}
		// Search by creation date
		creationDateStr := strconv.Itoa(artist.CreationDate)
		if strings.HasPrefix(creationDateStr, searchQuery) {
			results = append(results, SearchResult{
				Label:    creationDateStr + " - Creation Date of " + artist.Name,
				ID:       artist.ID,
				Category: "creation_date",
			})
		}
		// Search in Relations
		rel, err := getRelations(artist.ID)
		if err != nil {
			continue
		}
		for _, loc := range rel.SortedLocations {
			dates := rel.DatesLocations[loc]
			// Search by dates
			for _, date := range dates {
				if strings.Contains(date, searchQuery) {
					results = append(results, SearchResult{
						Label:    date + " - Concert date at " + loc + " for " + artist.Name,
						ID:       artist.ID,
						Category: "concert",
					})
				}
				// Search by location
				for _, part := range strings.Fields(strings.ToLower(loc)) {
					if strings.HasPrefix(part, searchQuery) {
						results = append(results, SearchResult{
							Label:    loc + " - Concert location on " + date + " for " + artist.Name,
							ID:       artist.ID,
							Category: "concert",
						})
					}
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

func FilterSearch(results []SearchResult, option string) []SearchResult {
	if option == "all" {
		return results
	}
	filtered := []SearchResult{}
	for _, result := range results {
		if result.Category == option {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
