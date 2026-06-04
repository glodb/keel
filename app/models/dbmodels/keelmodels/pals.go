package keelmodels

type Pal struct {
	Id                    string           `json:"id" bson:"id"`
	PalId                 string           `json:"palId" bson:"palId"`
	EventId               string           `json:"eventId" bson:"eventId"`
	OrganizerId           string           `json:"organizerId" bson:"organizerId"`
	OrganizerName         string           `json:"organizerName" bson:"organizerName"`
	OrganizerEmail        string           `json:"organizerEmail" bson:"organizerEmail"`
	OrganizerPhone        string           `json:"organizerPhone" bson:"organizerPhone"`
	OrganizerAvatar       string           `json:"organizerAvatar" bson:"organizerAvatar"`
	Type                  string           `json:"type" bson:"type"`
	Name                  string           `json:"name" bson:"name"`
	Email                 string           `json:"email" bson:"email"`
	VenueId               string           `json:"venueId" bson:"venueId"`
	VenueName             string           `json:"venueName" bson:"venueName"`
	FacilityId            string           `json:"facilityId" bson:"facilityId"`
	FacilityName          string           `json:"facilityName" bson:"facilityName"`
	SportsId              string           `json:"sportsId" bson:"sportsId"`
	LevelId               string           `json:"levelId" bson:"levelId"`
	BookingId             string           `json:"bookingId" bson:"bookingId"`
	AvatarUrls            []string         `json:"avatarUrls" bson:"avatarUrls"`
	InvitationSent        bool             `json:"invitationSent" bson:"invitationSent"`
	InvitationList        []PalInfo        `json:"invitationList" bson:"invitationList"`
	TotalSpots            int              `json:"totalSpots" bson:"totalSpots"`
	AcceptedList          []PalInfo        `json:"acceptedList" bson:"acceptedList"`
	Lat                   float64          `json:"lat" bson:"lat"`
	Lng                   float64          `json:"lng" bson:"lng"`
	Address               string           `json:"address" bson:"address"`
	City                  string           `json:"city" bson:"city"`
	Country               string           `json:"country" bson:"country"`
	PostalCode            string           `json:"postalCode" bson:"postalCode"`
	Phone                 string           `json:"phone" bson:"phone"`
	Whatsapp              string           `json:"whatsapp" bson:"whatsapp"`
	StartTime             int64            `json:"startTime" bson:"startTime"`
	EndTime               int64            `json:"endTime" bson:"endTime"`
	IsPrivate             bool             `json:"isPrivate" bson:"isPrivate"`
	IsPassed              bool             `json:"isPassed" bson:"isPassed"`
	CreatedAt             int64            `json:"createdAt" bson:"createdAt"`
	UpdatedAt             int64            `json:"updatedAt" bson:"updatedAt"`
	IsMeetingPoint        bool             `json:"isMeetingPoint" bson:"isMeetingPoint"`
	EnrollmentClosingTime int64            `json:"enrollmentClosingTime" bson:"enrollmentClosingTime"`
	Price                 []EventPriceItem `json:"price" bson:"price"`
	// Payment fields
	PaymentIntentId string  `json:"paymentIntentId,omitempty" bson:"paymentIntentId,omitempty"`
	PaymentStatus   string  `json:"paymentStatus,omitempty" bson:"paymentStatus,omitempty"` // "pending", "paid", "failed", "refunded"
	PerSpotPrice    float64 `json:"perSpotPrice,omitempty" bson:"perSpotPrice,omitempty"`
	TotalPrice      float64 `json:"totalPrice,omitempty" bson:"totalPrice,omitempty"`
	// Store payment intents for each participant (userId -> paymentIntentId)
	// ParticipantPaymentMethods stores (userId -> paymentMethodId) collected via SetupIntent on the client.
	ParticipantPaymentMethods map[string]string `json:"participantPaymentMethods,omitempty" bson:"participantPaymentMethods,omitempty"`
	// ParticipantPaymentIntents stores (userId -> paymentIntentId) created when the event becomes full and charges are attempted.
	ParticipantPaymentIntents map[string]string `json:"participantPaymentIntents,omitempty" bson:"participantPaymentIntents,omitempty"`
}

// GetQuery returns the query for finding this pal record
