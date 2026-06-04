package meilisearchmodels

import "github.com/glodb/keel/app/models/dbmodels/keelmodels"

const PalsIndexName = "pals"

// PalSearchDocument represents the pal document structure for Meilisearch
type PalSearchDocument struct {
	Id              string               `json:"id"`
	EventId         string               `json:"eventId"`
	OrganizerId     string               `json:"organizerId"`
	Type            string               `json:"type"`
	Name            string               `json:"name"`
	Email           string               `json:"email"`
	VenueId         string               `json:"venueId"`
	VenueName       string               `json:"venueName"`
	FacilityId      string               `json:"facilityId"`
	FacilityName    string               `json:"facilityName"`
	SportsId        string               `json:"sportsId"`
	LevelId         string               `json:"levelId"`
	BookingId       string               `json:"bookingId"`
	AvatarUrls      []string             `json:"avatarUrls"`
	InvitationSent  bool                 `json:"invitationSent"`
	TotalSpots      int                  `json:"totalSpots"`
	AcceptedUserIds []string             `json:"acceptedUserIds"` // Denormalized for filtering
	InvitedUserIds  []string             `json:"invitedUserIds"`  // Denormalized for filtering
	AcceptedList    []keelmodels.PalInfo `json:"acceptedList"`    // Full user details
	InvitationList  []keelmodels.PalInfo `json:"invitationList"`  // Full user details
	Lat             float64              `json:"lat"`
	Lng             float64              `json:"lng"`
	Geo             *GeoPoint            `json:"_geo,omitempty"` // For geo search
	Address         string               `json:"address"`
	City            string               `json:"city"`
	Country         string               `json:"country"`
	PostalCode      string               `json:"postalCode"`
	Phone           string               `json:"phone"`
	Whatsapp        string               `json:"whatsapp"`
	StartTime       int64                `json:"startTime"`
	EndTime         int64                `json:"endTime"`
	IsPrivate       bool                 `json:"isPrivate"`
	IsPassed        bool                 `json:"isPassed"`
	CreatedAt       int64                `json:"createdAt"`
	UpdatedAt       int64                `json:"updatedAt"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
