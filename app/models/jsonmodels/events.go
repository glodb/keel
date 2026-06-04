package jsonmodels

import "github.com/glodb/keel/app/models/dbmodels/keelmodels"

// CreateEventRequest models the request to create an event
type CreateEventRequest struct {
	EventName                string                         `json:"eventName" validate:"required"`
	EventId                  string                         `json:"eventId,omitempty"`
	SportsId                 string                         `json:"sportsId" validate:"required"`
	LevelId                  string                         `json:"levelId" validate:"required"`
	TotalSpots               int                            `json:"totalSpots" validate:"required,min=1"`
	BookingSpots             int                            `json:"bookingSpots"`
	Pals                     []string                       `json:"pals" validate:"max=20"`
	VenueId                  string                         `json:"venueId" validate:"required"`
	VenueName                string                         `json:"venueName"`
	FacilitiesInfo           []keelmodels.EventFacilityInfo `json:"facilitiesInfo"`
	Location                 keelmodels.EventLocation       `json:"location" validate:"required"`
	IsRepeatable             bool                           `json:"isRepeatable"`
	IsPrivate                bool                           `json:"isPrivate"`
	Price                    []keelmodels.EventPriceItem    `json:"price,omitempty"`
	TotalPrice               float64                        `json:"totalPrice,omitempty"`
	EnrollmentClosingSeconds int64                          `json:"enrollmentClosingSeconds,omitempty"`
	IsMeetingPoint           bool                           `json:"isMeetingPoint"`
	ValidFrom                int64                          `json:"validFrom,omitempty"`
	ValidUntil               int64                          `json:"validUntil,omitempty"`
}

// UpdateEventRequest models the request to update an event
type UpdateEventRequest struct {
	TotalSpots               int      `json:"totalSpots,omitempty" validate:"omitempty,min=1"`
	Pals                     []string `json:"pals,omitempty"`
	EnrollmentClosingSeconds int64    `json:"enrollmentClosingSeconds,omitempty"`
}

// GetEventRequest models the request to get an event by ID
type GetEventRequest struct {
	EventId string `json:"eventId" validate:"required" uri:"eventId"`
}

// DeleteEventRequest models the request to delete an event
type DeleteEventRequest struct {
	EventId string `json:"eventId" validate:"required" uri:"eventId"`
}
