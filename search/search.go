package search

import (
	"sort"
	"strconv"
	"strings"

	"groupie-tracker/models"
)

type SearchMethod int

const (
	MethodContains SearchMethod = 0
	MethodPrefix   SearchMethod = 1
)

type SearchResult struct {
	Label string
	ID    int
	Category string
	Method	 SearchMethod
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
					Method:   MethodPrefix,
				})
			} else if strings.Contains(part, searchQuery) {
				results = append(results, SearchResult{
					Label:    artist.Name + " - Artist/Band",
					ID:       artist.ID,
					Category: "artist",
					Method:   MethodContains,
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
						Method:   MethodPrefix,
					})
				} else if strings.Contains(part, searchQuery) {
					results = append(results, SearchResult{
						Label:    member + " - Member of " + artist.Name,
						ID:       artist.ID,
						Category: "member",
						Method:   MethodContains,
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
				Method:   MethodPrefix,
			})
		} else if strings.Contains(artist.FirstAlbum, searchQuery) {
			results = append(results, SearchResult{
				Label:    artist.FirstAlbum + " - First Album of " + artist.Name,
				ID:       artist.ID,
				Category: "first_album",
				Method:   MethodContains,
			})
		}
		// Search by creation date
		creationDateStr := strconv.Itoa(artist.CreationDate)
		if strings.HasPrefix(creationDateStr, searchQuery) {
			results = append(results, SearchResult{
				Label:    creationDateStr + " - Creation Date of " + artist.Name,
				ID:       artist.ID,
				Category: "creation_date",
				Method:   MethodPrefix,
			})
		} else if strings.Contains(creationDateStr, searchQuery) {
			results = append(results, SearchResult{
				Label:    creationDateStr + " - Creation Date of " + artist.Name,
				ID:       artist.ID,
				Category: "creation_date",
				Method:   MethodContains,
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
				if strings.HasPrefix(date, searchQuery) {
					results = append(results, SearchResult{
						Label:    date + " - Concert date at " + loc + " for " + artist.Name,
						ID:       artist.ID,
						Category: "concert",
						Method:   MethodPrefix,
					})
				} else if strings.Contains(date, searchQuery) {
					results = append(results, SearchResult{
						Label:    date + " - Concert date at " + loc + " for " + artist.Name,
						ID:       artist.ID,
						Category: "concert",
						Method:   MethodContains,
					})
				}
				// Search by location
				for _, part := range strings.Fields(strings.ToLower(normalize(loc))) {
					if strings.HasPrefix(part, normalize(searchQuery)) {
						results = append(results, SearchResult{
							Label:    loc + " - Concert location on " + date + " for " + artist.Name,
							ID:       artist.ID,
							Category: "concert",
							Method:   MethodPrefix,
						})
					} else if strings.Contains(part, normalize(searchQuery)) {
						results = append(results, SearchResult{
							Label:    loc + " - Concert location on " + date + " for " + artist.Name,
							ID:       artist.ID,
							Category: "concert",
							Method:   MethodContains,
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
	s = strings.ReplaceAll(s, ":", "")
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

// ParseQuery splits the search query into tokens
// ex. "Pink Floyd" -> {"pink" "floyd"}
func ParseQuery(query string) []string {
	return strings.Fields(strings.ToLower(query))
}

// SortResults sorts the search results by method (prefix matches before contains matches)
func SortResults(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Method > results[j].Method
	})
}

// MatchResults returns only the results whose IDs appear in all token results
func MatchResults(tokenResults [][]SearchResult) []SearchResult {
	if len(tokenResults) == 0 {
		return nil
	}
	// Count how many token-groups each ID appears in
	idTokenCount := make(map[int]int)
	for _, results := range tokenResults {
		seen := make(map[int]bool)
		for _, r := range results {
			if seen[r.ID] {
				continue
			}
			seen[r.ID] = true
			idTokenCount[r.ID]++
		}
	}
	required := len(tokenResults)
	// Determine which IDs appear in ALL token groups
	validIDs := make(map[int]bool)
	for id, count := range idTokenCount {
		if count == required {
			validIDs[id] = true
		}
	}
	// Collect ALL results belonging to valid IDs
	matched := []SearchResult{}
	for _, results := range tokenResults {
		for _, r := range results {
			if validIDs[r.ID] {
				matched = append(matched, r)
			}
		}
	}
	return matched
}

// RemoveDuplicates removes duplicate search results based on Label
func RemoveDuplicates(results []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	unique := []SearchResult{}
	for _, r := range results {
		if !seen[r.Label] {
			seen[r.Label] = true
			unique = append(unique, r)
		}
	}
	return unique
}

// Search performs a full search based on the query string.
// It splits the query into tokens, searches for each token, matches results that appear in all tokens,
// sorts the results, and removes duplicates.
func Search(query string, artists []models.Artists, getRelations func(int) (*models.Relations, error)) []SearchResult {
	// Tokenize the query
	tokens := ParseQuery(query)
	if len(tokens) == 1 {
		// Single token search
		return SearchAll(tokens[0], artists, getRelations)
	}
	// Multi-token search
	resultsPerToken := [][]SearchResult{}
	for _, token := range tokens {
		tokenResults := SearchAll(token, artists, getRelations)
		resultsPerToken = append(resultsPerToken, tokenResults)
	}
	// Match results that appear in all tokens
	results := MatchResults(resultsPerToken)
	// Sort results
	SortResults(results)
	return RemoveDuplicates(results)
}
