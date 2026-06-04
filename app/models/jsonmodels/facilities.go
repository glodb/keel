package jsonmodels

import "github.com/glodb/keel/app/models/dbmodels/keelmodels"

// CreateFacilityRequest models the request to create a facility
type CreateFacilityRequest struct {
	VenueId           string                     `json:"venueId" validate:"required"`
	Name              string                     `json:"name" validate:"required"`
	Description       string                     `json:"description,omitempty"`
	Image             string                     `json:"image,omitempty"`
	Icon              string                     `json:"icon,omitempty"`
	Price             float64                    `json:"price,omitempty"`
	SportId           string                     `json:"sportId" validate:"required"`
	IsRepeatable      bool                       `json:"isRepeatable" validate:"required"`
	WeeklyPattern     *keelmodels.WeeklySchedule `json:"weeklyPattern,omitempty"`
	SpecificDates     []keelmodels.DailySchedule `json:"specificDates,omitempty"`
	TotalSpotsPerSlot int                        `json:"totalSpotsPerSlot" validate:"required,min=1"`
	ValidFrom         int64                      `json:"validFrom" validate:"required"`
	ValidUntil        int64                      `json:"validUntil,omitempty"` // 0 means no expiry
}

// UpdateFacilityRequest models the request to update a facility
type UpdateFacilityRequest struct {
	FacilityId        string                     `json:"facilityId" validate:"required"`
	Name              string                     `json:"name,omitempty"`
	Description       string                     `json:"description,omitempty"`
	Image             string                     `json:"image,omitempty"`
	Icon              string                     `json:"icon,omitempty"`
	Price             float64                    `json:"price,omitempty"`
	SportId           string                     `json:"sportId,omitempty"`
	PatternId         string                     `json:"patternId" validate:"required" uri:"patternId"`
	WeeklyPattern     *keelmodels.WeeklySchedule `json:"weeklyPattern,omitempty"`
	SpecificDates     []keelmodels.DailySchedule `json:"specificDates,omitempty"`
	TotalSpotsPerSlot int                        `json:"totalSpotsPerSlot,omitempty" validate:"omitempty,min=1"`
	ValidFrom         int64                      `json:"validFrom,omitempty"`
	ValidUntil        int64                      `json:"validUntil,omitempty"`
	IsRepeatable      bool                       `json:"isRepeatable,omitempty"`
}

// GetFacilitiesRequest models the request to get facilities for a venue
type GetFacilitiesRequest struct {
	VenueId string `json:"venueId" validate:"required" uri:"venueId"`
}

// GetAvailabilityPatternsRequest models the request to get availability patterns
type GetAvailabilityPatternsRequest struct {
	FacilityId string `json:"facilityId" validate:"required" uri:"facilityId"`
}

// CreateFacilityBookingRequest models the request to create a booking
// Follows the same pattern as facility availability using IsRepeatable, WeeklyPattern, and SpecificDates
type CreateFacilityBookingRequest struct {
	FacilityId    string                     `json:"facilityId" validate:"required"`
	EventId       string                     `json:"eventId" validate:"required"`
	SpotsBooked   int                        `json:"spotsBooked" validate:"required,min=1"`
	IsRepeatable  bool                       `json:"isRepeatable"`                  // true for recurring, false for specific dates
	WeeklyPattern *keelmodels.WeeklySchedule `json:"weeklyPattern,omitempty"`       // For recurring bookings
	SpecificDates []keelmodels.DailySchedule `json:"specificDates,omitempty"`       // For specific dates/one-time bookings
	ValidFrom     int64                      `json:"validFrom" validate:"required"` // Unix timestamp - when booking starts
	ValidUntil    int64                      `json:"validUntil,omitempty"`          // Unix timestamp - when booking ends (0 = indefinite for recurring)
}

// CancelBookingRequest models the request to cancel a booking
type CancelBookingRequest struct {
	BookingId string `json:"bookingId" validate:"required" uri:"bookingId"`
}

// CheckAvailabilityRequest models the request to check facility availability
type CheckAvailabilityRequest struct {
	VenueId      string                     `json:"venueId" validate:"required"`
	SportId      string                     `json:"sportId,omitempty"`
	IsRepeatable bool                       `json:"isRepeatable"`
	Schedule     *keelmodels.WeeklySchedule `json:"schedule,omitempty"` // For recurring
	Dates        []keelmodels.DailySchedule `json:"dates,omitempty"`    // For specific dates/one-time
	SpotsNeeded  int                        `json:"spotsNeeded" validate:"required,min=1"`
}

// GetFacilitiesEventsRequest models the request to get events for multiple facilities
type GetFacilitiesEventsRequest struct {
	FacilityIds []string `json:"facilityIds" validate:"required,min=1"`
	Limit       int      `json:"limit" validate:"omitempty,min=1,max=100"`
	Page        int      `json:"page" validate:"omitempty,min=1"`
	Sort        int      `json:"sort" validate:"omitempty,min=1,max=100"`
	SortValue   string   `json:"sortValue" validate:"omitempty,min=1,max=100"`
	RangeKey    string   `json:"rangeKey" validate:"omitempty,min=1,max=100"`
	StartTime   int64    `json:"startTime" validate:"omitempty,min=1"`
	EndTime     int64    `json:"endTime" validate:"omitempty,min=1"`
}

func (u *GetFacilitiesEventsRequest) SetPage(page int) {
	u.Page = page
}

func (u *GetFacilitiesEventsRequest) SetLimit(limit int) {
	u.Limit = limit
}

func (u *GetFacilitiesEventsRequest) SetSort(sort int) {
	u.Sort = sort
}

func (u *GetFacilitiesEventsRequest) SetSortValue(sortValue string) {
	u.SortValue = sortValue
}

func (u *GetFacilitiesEventsRequest) SetRangeKey(rangeKey string) {
	u.RangeKey = rangeKey
}

func (u *GetFacilitiesEventsRequest) SetStartTime(startTime int64) {
	u.StartTime = startTime
}

func (u *GetFacilitiesEventsRequest) SetEndTime(endTime int64) {
	u.EndTime = endTime
}

func (u GetFacilitiesEventsRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	u.Page = page
	u.Limit = limit
	u.Sort = sort
	u.SortValue = sortValue
	u.RangeKey = rangeKey
	u.StartTime = startTime
	u.EndTime = endTime
}

func (u GetFacilitiesEventsRequest) GetPage() int {
	return u.Page
}

func (u GetFacilitiesEventsRequest) GetLimit() int {
	return u.Limit
}

func (u GetFacilitiesEventsRequest) GetSort() int {
	return u.Sort
}

func (u GetFacilitiesEventsRequest) GetSortValue() string {
	return u.SortValue
}

func (u GetFacilitiesEventsRequest) GetRangeKey() string {
	return u.RangeKey
}

func (u GetFacilitiesEventsRequest) GetStartTime() int64 {
	return u.StartTime
}

func (u GetFacilitiesEventsRequest) GetEndTime() int64 {
	return u.EndTime
}

// FacilityAvailabilityResponse models the response for facility availability
type FacilityAvailabilityResponse struct {
	FacilityId       string   `json:"facilityId"`
	FacilityName     string   `json:"facilityName"`
	AvailableSpots   int      `json:"availableSpots"`
	TotalSpots       int      `json:"totalSpots"`
	IsFullyAvailable bool     `json:"isFullyAvailable"`           // All requested times available
	ConflictingDates []string `json:"conflictingDates,omitempty"` // Dates where conflicts exist
}

// DetailedTimeSlotAvailability represents availability for a specific time slot
type DetailedTimeSlotAvailability struct {
	Day            string              `json:"day"`              // "monday", "tuesday" or date "2024-01-15"
	TimeSlot       keelmodels.TimeSlot `json:"timeSlot"`         // {startTime, endTime}
	Available      bool                `json:"available"`        // Is this slot available?
	AvailableSpots int                 `json:"availableSpots"`   // How many spots available
	TotalSpots     int                 `json:"totalSpots"`       // Total spots in facility
	Reason         string              `json:"reason,omitempty"` // Reason if not available
}

// DetailedFacilityAvailability represents detailed availability for a facility
type DetailedFacilityAvailability struct {
	FacilityId   string                         `json:"facilityId"`
	FacilityName string                         `json:"facilityName"`
	Price        float64                        `json:"price"` // Facility price per booking
	TotalSpots   int                            `json:"totalSpots"`
	Availability []DetailedTimeSlotAvailability `json:"availability"`
}

type VenueFacilitiesSport struct {
	VenueId     string `json:"venueId"`
	FacilityIds string `json:"facilityIds"`
	SportId     string `json:"sportId"`
}

// CheckUserAvailabilityRequest models the request to check user's personal availability
// Used when creating meeting places (not booking venues)
type CheckUserAvailabilityRequest struct {
	StartDate string `json:"startDate,omitempty"` // Optional: YYYY-MM-DD format, defaults to today
	EndDate   string `json:"endDate,omitempty"`   // Optional: YYYY-MM-DD format, defaults to 30 days from start
	PalUserId string `json:"palUserId,omitempty"` // Optional: If provided, match slots with this pal's availability
}

// UserAvailabilityResponse represents the response with user's available time slots
type UserAvailabilityResponse struct {
	IsRepeatable bool                           `json:"isRepeatable"` // true if weekly pattern, false if specific dates
	Availability []DetailedTimeSlotAvailability `json:"availability"` // 1-hour brackets of free time
}
