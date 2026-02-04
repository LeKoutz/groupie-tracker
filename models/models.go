package models

type Artists struct {
	ID				int		 `json:"id"`
	Image			string	 `json:"image"`
	Name			string	 `json:"name"`
	Members			[]string `json:"members"`
	CreationDate	int		 `json:"creationDate"`
	FirstAlbum		string	 `json:"firstAlbum"`
	Locations		string	 `json:"locations"`
	ConcertDates	string	 `json:"concertDates"`
	Relations		string	 `json:"relations"`
}

type LocationsIndex struct {
	Index 			[]Locations	 `json:"index"`
}

type Locations struct {
	ID				int		 `json:"id"`
	Locations 		[]string `json:"locations"`
	Dates			string	 `json:"dates"`
}

type DatesIndex struct {
	Index 			[]Dates	 `json:"index"`
}

type Dates struct {
	ID				int		 `json:"id"`
	ConcertDates	[]string `json:"dates"`
}

type RelationIndex struct {
	Index 			[]Relations `json:"index"`
}

type Relations struct {
	ID				int		 			 `json:"id"`
	DatesLocations	map[string][]string	 `json:"datesLocations"`
	SortedLocations []string
}

type ArtistDetails struct {
	Artist		Artists
	Locations	Locations
	Dates		Dates
	Relations	Relations
	MapData map[string]Coordinates
}

// struct to store latitude and longitude
type Coordinates struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}