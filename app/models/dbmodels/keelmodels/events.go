package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// EventPriceItem represents a price item in an event
type EventPriceItem struct {
	Type  string  `json:"type" bson:"type"`
	Name  string  `json:"name" bson:"name"`
	Price float64 `json:"price" bson:"price"`
}

// EventLocation represents location for an event (either venueId or LatLng)
type EventLocation struct {
	VenueId string  `json:"venueId,omitempty" bson:"venueId,omitempty" validate:"omitempty,required_without_all=Lat Lng"`
	Lat     float64 `json:"lat,omitempty" bson:"lat,omitempty" validate:"omitempty,required_with=Lng,min=-90,max=90"`
	Lng     float64 `json:"lng,omitempty" bson:"lng,omitempty" validate:"omitempty,required_with=Lat,min=-180,max=180"`
	Address string  `json:"address,omitempty" bson:"address,omitempty"`
	City    string  `json:"city,omitempty" bson:"city,omitempty"`
	Country string  `json:"country,omitempty" bson:"country,omitempty"`
}

// Event represents an event entity in the database
type Event struct {
	UserId                   string              `json:"userId" bson:"userId"`
	LastEndDate              int64               `json:"lastEndDate" bson:"lastEndDate"`
	LastProcessedAt          int64               `json:"lastProcessedAt" bson:"lastProcessedAt"`
	EventId                  string              `json:"eventId" bson:"eventId"`
	EventName                string              `json:"eventName" bson:"eventName"`
	VenueId                  string              `json:"venueId" bson:"venueId"`
	VenueName                string              `json:"venueName" bson:"venueName"`
	OrganizerName            string              `json:"organizerName" bson:"organizerName"`
	OrganizerEmail           string              `json:"organizerEmail" bson:"organizerEmail"`
	OrganizerPhone           string              `json:"organizerPhone" bson:"organizerPhone"`
	OrganizerAvatar          string              `json:"organizerAvatar" bson:"organizerAvatar"`
	SportsId                 string              `json:"sportsId" bson:"sportsId"`
	LevelId                  string              `json:"levelId" bson:"levelId"`
	TotalSpots               int                 `json:"totalSpots" bson:"totalSpots"`
	Pals                     []PalInfo           `json:"pals" bson:"pals"` // Array of PalInfo
	Location                 EventLocation       `json:"location" bson:"location"`
	IsRepeatable             bool                `json:"isRepeatable" bson:"isRepeatable"`
	IsPrivate                bool                `json:"isPrivate" bson:"isPrivate"`
	Price                    []EventPriceItem    `json:"price,omitempty" bson:"price,omitempty"`
	OrganizerId              string              `json:"organizerId" bson:"organizerId"` // UserId of the event creator
	IsUpdate                 bool                `json:"isUpdate" bson:"isUpdate"`
	ValidFrom                int64               `json:"validFrom" bson:"validFrom"`
	ValidUntil               int64               `json:"validUntil" bson:"validUntil"`
	FacilitiesInfo           []EventFacilityInfo `json:"facilitiesInfo" bson:"facilitiesInfo"`
	IsMeetingPoint           bool                `json:"isMeetingPoint" bson:"isMeetingPoint"`
	EnrollmentClosingSeconds int64               `json:"enrollmentClosingSeconds" bson:"enrollmentClosingSeconds"`
	CreatedAt                int64               `json:"createdAt" bson:"createdAt"`
	UpdatedAt                int64               `json:"updatedAt" bson:"updatedAt"`
	// Payment fields
	PaymentIntentId string  `json:"paymentIntentId,omitempty" bson:"paymentIntentId,omitempty"`
	PaymentStatus   string  `json:"paymentStatus,omitempty" bson:"paymentStatus,omitempty"` // "pending", "paid", "failed", "refunded"
	PerSpotPrice    float64 `json:"perSpotPrice,omitempty" bson:"perSpotPrice,omitempty"`
	TotalPrice      float64 `json:"totalPrice,omitempty" bson:"totalPrice,omitempty"`
}

// GetQuery returns the query for finding this event record
func (e *Event) GetQuery() map[string]interface{} {
	return map[string]interface{}{"eventId": e.EventId}
}

// GetUpdate returns the update data for this event record
func (e *Event) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"eventName":  e.EventName,
		"sportsId":   e.SportsId,
		"levelId":    e.LevelId,
		"totalSpots": e.TotalSpots,
		"pals":       e.Pals,
		"location":   e.Location,
		"isPrivate":  e.IsPrivate,
		"price":      e.Price,
		"updatedAt":  time.Now().Unix(),
	}
}

// MapData maps the provided data to the Event struct
func (e *Event) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, e)
}

// EncodeRedisData encodes the Event data for Redis storage
func (e *Event) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(e)
	return bytes
}

// DecodeRedisData decodes the Event data from Redis storage
func (e *Event) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, e)
}
