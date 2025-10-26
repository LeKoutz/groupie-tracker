package services

import (
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
)


func GetArtistByID(id int) (*models.Artists, error) {
	for i := range api.All_Artists {
		if api.All_Artists[i].ID == id {
			return &api.All_Artists[i], nil
		}
	}
	return nil, fmt.Errorf("Error: Artist ID %d not found", id)
}

func GetLocationsByID(id int) (*models.Locations, error) {
	for i := range api.All_Locations {
		if api.All_Locations[i].ID == id {
			return &api.All_Locations[i], nil
		}
	}
	return nil, fmt.Errorf("Error: No locations found for ID %d", id)
}

func GetDatesByID(id int) (*models.Dates, error) {
	for i := range api.All_Dates {
		if api.All_Dates[i].ID == id {
			return &api.All_Dates[i], nil
		}
	}
	return nil, fmt.Errorf("Error: No dates found for ID %d", id)
}

func GetRelationsByID(id int) (*models.Relations, error) {
	for i := range api.All_Relations {
		if api.All_Relations[i].ID == id {
			return &api.All_Relations[i], nil
		}
	}
	return nil, fmt.Errorf("Error: No relations found for ID %d", id)
}
