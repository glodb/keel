package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// FacilityBooking represents an actual booking (recurring or specific dates)
// Follows the same pattern as events and user schedules using IsRepeatable, WeeklySchedule, and DailySchedule
type FacilityBooking struct {
	BookingId  string `json:"bookingId" bson:"bookingId"`
	FacilityId string `json:"facilityId" bson:"facilityId"`
	VenueId    string `json:"venueId" bson:"venueId"`
	EventId    string `json:"eventId" bson:"eventId"` // Reference to the event
	UserId     string `json:"userId" bson:"userId"`   // User who made the booking

	SpotsBooked int `json:"spotsBooked" bson:"spotsBooked"` // Number of spots booked

	// Following the existing pattern structure used in events and schedules
	IsRepeatable  bool            `json:"isRepeatable" bson:"isRepeatable"`                       // true for recurring, false for specific dates
	WeeklyPattern *WeeklySchedule `json:"weeklyPattern,omitempty" bson:"weeklyPattern,omitempty"` // For recurring bookings
	SpecificDates []DailySchedule `json:"specificDates,omitempty" bson:"specificDates,omitempty"` // For specific dates/one-time bookings

	// Validity period
	ValidFrom  int64 `json:"validFrom" bson:"validFrom"`                       // Unix timestamp - when booking starts
	ValidUntil int64 `json:"validUntil,omitempty" bson:"validUntil,omitempty"` // Unix timestamp - when booking ends (0 = indefinite for recurring)

	// Metadata
	Status    string `json:"status" bson:"status"` // "ACTIVE", "CANCELLED", "COMPLETED"
	CreatedAt int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string `json:"createdBy" bson:"createdBy"`
	UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
}

// GetQuery returns the query for finding this booking record
func (f *FacilityBooking) GetQuery() map[string]interface{} {
	return map[string]interface{}{"bookingId": f.BookingId}
}

// GetUpdate returns the update data for this booking record
func (f *FacilityBooking) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"status":    f.Status,
		"updatedBy": f.UpdatedBy,
		"updatedAt": time.Now().Unix(),
	}
}

// MapData maps the provided data to the FacilityBooking struct
func (f *FacilityBooking) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, f)
}

// EncodeRedisData encodes the FacilityBooking data for Redis storage
func (f *FacilityBooking) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(f)
	return bytes
}

// DecodeRedisData decodes the FacilityBooking data from Redis storage
func (f *FacilityBooking) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, f)
}
