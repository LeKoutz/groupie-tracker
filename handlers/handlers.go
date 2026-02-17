package handlers

import (
	"encoding/json"
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"groupie-tracker/services"
	"groupie-tracker/search"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

// Parse templates once before server startup
var (
	index_tmpl   = template.Must(template.ParseFiles("templates/index.html"))
	artist_tmpl  = template.Must(template.ParseFiles("templates/artist_detail.html"))
	error_tmpl   = template.Must(template.ParseFiles("templates/error.html"))
	loading_tmpl = template.Must(template.ParseFiles("templates/loading.html"))
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), "Please check the resource URL and try again.")
		return
	}
	if r.Method != http.MethodGet {
		HandleErrors(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), "This request method is not supported for the requested resource. Use GET request instead.")
		return
	}
	if api.GetLoadingStatus().IsLoading {
		http.Redirect(w, r, "/loading?requested="+url.QueryEscape(r.URL.Path), http.StatusSeeOther)
		return
	} else if api.GetLoadingStatus().HasFailed {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to load the data. Please try again later.")
		return
	}
	query := r.URL.Query().Get("search")
	var SearchResults []search.SearchResult
	if query != "" {
		SearchResults = search.Search(query, api.All_Artists, services.GetRelationsByID)
	}
	category := r.URL.Query().Get("category")
	if category != "" && category != "all" {
		SearchResults = search.FilterSearch(SearchResults, category)
	}
	data := struct {
		Artists       []models.Artists
		SearchQuery   string
		SearchResults []search.SearchResult
		NoResults	  bool
	}{
		Artists:       api.All_Artists,
		SearchQuery:   query,
		SearchResults: SearchResults,
		NoResults:     false,
	}
	// If query exists and SearchResults != empty, show search results only
	if query != "" && len(SearchResults) > 0 {
		data.Artists = []models.Artists{}
		for _, result := range SearchResults {
			artist, err := services.GetArtistByID(result.ID)
			// Append artist to data.Artists if not already appended
			if err == nil && !services.ArtistExistsInList(data.Artists, artist) {
				data.Artists = append(data.Artists, *artist)
			}
		}
	} else if query != "" && len(SearchResults) == 0 {
		data.SearchResults = []search.SearchResult{}
		data.NoResults = true
	}
	if err := index_tmpl.Execute(w, data); err != nil {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to complete your request. Please try again later")
		return
	}
}

func ArtistDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/artist/") {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), "Please check the resource URL and try again.")
		return
	}
	if r.Method != http.MethodGet {
		HandleErrors(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), "This request method is not supported for the requested resource. Use GET request instead.")
		return
	}
	if api.GetLoadingStatus().IsLoading {
		http.Redirect(w, r, "/loading?requested="+ url.QueryEscape(r.URL.Path), http.StatusSeeOther)
		return
	} else if api.GetLoadingStatus().HasFailed {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to load the data. Please try again later.")
		return
	}
	artist_ID, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/artist/"))

	artist, err := services.GetArtistByID(artist_ID)
	if err != nil {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err.Error())
		return
	}
	locations, err := services.GetLocationsByID(artist_ID)
	if err != nil {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err.Error())
		return
	}
	dates, err := services.GetDatesByID(artist_ID)
	if err != nil {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err.Error())
		return
	}
	relations, err := services.GetRelationsByID(artist_ID)
	if err != nil {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), err.Error())
		return
	}
	mapData := services.Geocode(relations.SortedLocations)
	data := models.ArtistDetails{
		Artist:    *artist,
		Locations: *locations,
		Dates:     *dates,
		Relations: *relations,
		MapData:   mapData,
	}
	if err := artist_tmpl.Execute(w, data); err != nil {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to complete your request. Please try again later")
		return
	}
}

func ResourcesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HandleErrors(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), "This request method is not supported for the requested resource. Use GET request instead.")
		return
	}
	if strings.HasSuffix(r.URL.Path, "/") {
		HandleErrors(w, http.StatusNotFound, http.StatusText(http.StatusNotFound), "Please check the resource URL and try again.")
		return
	}
	filePath := strings.TrimPrefix(r.URL.Path, "/static")
	http.ServeFile(w, r, filepath.Join("static", filePath))
}

// HandleErrors renders an error page with the given status code, error message, and response.
func HandleErrors(w http.ResponseWriter, statusCode int, message, response string) {
	errorData := struct {
		StatusCode int
		Message    string
		Response   string
	}{
		StatusCode: statusCode,
		Message:    http.StatusText(statusCode),
		Response:   response,
	}
	w.WriteHeader(statusCode)
	if err := error_tmpl.Execute(w, errorData); err != nil {
		http.Error(w, fmt.Sprintf("Error %d: %s %s", statusCode, message, response), statusCode)
		return
	}
}

func LoadingHandler(w http.ResponseWriter, r *http.Request) {
	status := api.GetLoadingStatus()
	requestedURL := r.URL.Query().Get("requested")
	if requestedURL == "" {
		requestedURL = "/"
	}
	if status.IsLoaded {
		http.Redirect(w, r, requestedURL, http.StatusSeeOther)
		return
	}
	if status.HasFailed {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to load the data. Please try again later.")
		return
	} else {
		w.Header().Set("Refresh", "1; url="+requestedURL)
		data := struct {
			Message string
		}{
			Message: "Loading data...",
		}
		if err := loading_tmpl.Execute(w, data); err != nil {
			HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to complete your request. Please try again later")
			return
		}
	}
}

// SearchHandler returns search results in JSON format based on the query parameter.
// It is used for Javascript-based search autocomplete functionality.
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HandleErrors(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), "This request method is not supported for the requested resource. Use GET request instead.")
		return
	}

	// Parse filter parameters
	filters := parseFilterParams(r)

	query := r.URL.Query().Get("search")

	// We will fetch the filtered list of artists first
	filteredArtists := services.FilterArtists(api.All_Artists, api.All_Locations, filters)

	var SearchResults []search.SearchResult

	if query != "" {
		// If there is a text query, search within the filtered artists
		SearchResults = search.Search(query, filteredArtists, services.GetRelationsByID)

		category := r.URL.Query().Get("category")
		if category != "" && category != "all" {
			SearchResults = search.FilterSearch(SearchResults, category)
		}
	} else {

		// Converting models.Artists to search.SearchResult
		for _, artist := range filteredArtists {
			SearchResults = append(SearchResults, search.SearchResult{
				Label:    artist.Name,
				ID:       artist.ID,
				Category: "artist",
				Method:   search.MethodContains,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SearchResults)
}

func parseFilterParams(r *http.Request) models.FilterParameters {
	params := r.URL.Query()

	// Helper to parse int with default
	parseInt := func(key string, defaultValue int) int {
		if val := params.Get(key); val != "" {
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
		return defaultValue
	}

	// Set sensible defaults for "Max" values to ensure we include everything if not specified.

	return models.FilterParameters{
		MinCreationDate:   parseInt("min_creation_date", 0),
		MaxCreationDate:   parseInt("max_creation_date", 3000),
		MinFirstAlbumYear: parseInt("min_first_album_year", 0),
		MaxFirstAlbumYear: parseInt("max_first_album_year", 3000),
		MinMembers:        parseInt("min_members", 0),
		MaxMembers:        parseInt("max_members", 100),
		SelectedLocations: func() []string {
			if locs := params.Get("locations"); locs != "" {
				return strings.Split(locs, ",")
			}
			return nil
		}(),
	}
}
