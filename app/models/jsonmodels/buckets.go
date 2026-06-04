package jsonmodels

type BucketData struct {
	EventId               string   `json:"eventId"`
	Name                  string   `json:"name"`
	PalId                 string   `json:"palId"`
	VenueId               string   `json:"venueId"`
	City                  string   `json:"city"`
	Address               string   `json:"address"`
	SportsId              string   `json:"sportsId"`
	StartTime             int64    `json:"startTime"`
	EndTime               int64    `json:"endTime"`
	EnrollmentClosingTime int64    `json:"enrollmentClosingTime"`
	LevelId               string   `json:"levelId"`
	Lat                   float64  `json:"lat"`
	SpotsLeft             int      `json:"spotsLeft"`
	Lng                   float64  `json:"lng"`
	Type                  string   `json:"type"`
	Distance              float64  `json:"distance"`
	AvatarUrls            []string `json:"avatarUrls"`
}

type BucketDataForVenue struct {
	VenueId      string  `json:"venueId"`
	EventName    string  `json:"eventName"`
	PalId        string  `json:"palId"`
	SportsId     string  `json:"sportsId"`
	StartTime    int64   `json:"startTime"`
	EndTime      int64   `json:"endTime"`
	SpotsLeft    int     `json:"spotsLeft"`
	LevelId      string  `json:"levelId"`
	FacilityId   string  `json:"facilityId"`
	VenueName    string  `json:"venueName"`
	FacilityName string  `json:"facilityName"`
	OrganizerId  string  `json:"organizerId"`
	Price        float64 `json:"price"`
}
