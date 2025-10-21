package api

import (
	"encoding/json"
	"net/http"
	"groupie-tracker/models"
)

func FetchArtists() []models.Artist {
	resp, _ := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	defer resp.Body.Close()

	var artists []models.Artist
	json.NewDecoder(resp.Body).Decode(&artists)
	return artists
}

func FetchLocationsByURL(url string) []models.Locations {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var concert_locations []models.Locations
	json.NewDecoder(resp.Body).Decode(&concert_locations)
	return concert_locations
}

func FetchDatesByURL(url string) []models.Dates {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var concert_dates []models.Dates
	json.NewDecoder(resp.Body).Decode(&concert_dates)
	return concert_dates
}

func FetchRelationsByURL(url string) []models.Relation {
	resp, _ := http.Get(url)
	defer resp.Body.Close()

	var relations []models.Relation
	json.NewDecoder(resp.Body).Decode(&relations)
	return relations
}
