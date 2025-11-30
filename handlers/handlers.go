package handlers

import (
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"groupie-tracker/services"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Parse templates once before server startup
var (
	index_tmpl   = template.Must(template.ParseFiles("templates/index.html"))
	artist_tmpl  = template.Must(template.ParseFiles("templates/artist_detail.html"))
	error_tmpl   = template.Must(template.ParseFiles("templates/error.html"))
	loading_tmpl = template.Must(template.ParseFiles("templates/loading.html"))
)

type DateLocation struct {
	Date            string
	Location        string
	LocationDisplay string
}

func formatLocationDisplay(raw string) string {
	// split "city-country" on the first "-"
	parts := strings.SplitN(raw, "-", 2)

	cityRaw := parts[0]
	countryRaw := ""
	if len(parts) > 1 {
		countryRaw = parts[1]
	}

	// replace "_" with spaces
	city := strings.ReplaceAll(cityRaw, "_", " ")
	country := strings.ReplaceAll(countryRaw, "_", " ")

	// basic title case
	city = strings.Title(city)
	country = strings.Title(country)

	// special cases for country
	switch strings.ToLower(countryRaw) {
	case "usa":
		country = "USA"
	case "uk":
		country = "UK"
	}

	if country == "" {
		return city
	}
	return city + ", " + country
}

func parseAPIDate(dateString string) time.Time {
	// API format: "dd-mm-yyyy"
	parsed, err := time.Parse("02-01-2006", dateString)
	if err != nil {
		return time.Time{} // zero value if something is wrong
	}
	return parsed
}

func BuildDateLocations(datesLocations map[string][]string) []DateLocation {
	result := make([]DateLocation, 0)

	for location, dates := range datesLocations {
		for _, date := range dates {
			result = append(result, DateLocation{
				Date:            date,
				Location:        location,
				LocationDisplay: formatLocationDisplay(location),
			})
		}
	}

	// sort newest → oldest
	sort.Slice(result, func(i, j int) bool {
		di := parseAPIDate(result[i].Date)
		dj := parseAPIDate(result[j].Date)

		// newer date should come first
		return di.After(dj)
	})

	return result
}

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
		http.Redirect(w, r, "/loading/", http.StatusSeeOther)
		return
	}
	data := struct {
		Artists []models.Artists
	}{
		Artists: api.All_Artists,
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

	artistDetails := models.ArtistDetails{
		Artist:    *artist,
		Locations: *locations,
		Dates:     *dates,
		Relations: *relations,
	}

	dateLocations := BuildDateLocations(artistDetails.Relations.DatesLocations)

	data := struct {
		ArtistDetails models.ArtistDetails
		DateLocations []DateLocation
	}{
		ArtistDetails: artistDetails,
		DateLocations: dateLocations,
	}

	if err := artist_tmpl.Execute(w, data); err != nil {
		HandleErrors(
			w,
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
			"The server was unable to complete your request. Please try again later",
		)
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
	if status.IsLoaded {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if status.HasFailed {
		HandleErrors(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), "The server was unable to load the data. Please try again later.")
		return
	} else {
		w.Header().Set("Refresh", "1; url=/loading")
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
