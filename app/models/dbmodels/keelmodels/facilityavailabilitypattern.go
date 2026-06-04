package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// FacilityAvailabilityPattern represents the availability rules for a facility
// This stores patterns, not expanded slots
type FacilityAvailabilityPattern struct {
	PatternId    string `json:"patternId" bson:"patternId"`
	FacilityId   string `json:"facilityId" bson:"facilityId"`
	VenueId      string `json:"venueId" bson:"venueId"`
	IsRepeatable bool   `json:"isRepeatable" bson:"isRepeatable"`

	// For WEEKLY_RECURRING pattern
	WeeklyPattern *WeeklySchedule `json:"weeklyPattern,omitempty" bson:"weeklyPattern,omitempty"`

	// For SPECIFIC_DATES pattern
	SpecificDates []DailySchedule `json:"specificDates,omitempty" bson:"specificDates,omitempty"`

	// Spots configuration
	TotalSpotsPerSlot int `json:"totalSpotsPerSlot" bson:"totalSpotsPerSlot"` // Total spots available per time slot

	// Validity period
	ValidFrom  int64 `json:"validFrom" bson:"validFrom"`   // Unix timestamp - when this pattern becomes active
	ValidUntil int64 `json:"validUntil" bson:"validUntil"` // Unix timestamp - when this pattern expires (0 = no expiry)

	// Metadata
	IsActive  bool   `json:"isActive" bson:"isActive"`
	CreatedAt int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" bson:"updatedAt"`
	CreatedBy string `json:"createdBy" bson:"createdBy"`
	UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
}

// GetQuery returns the query for finding this pattern record
func (f *FacilityAvailabilityPattern) GetQuery() map[string]interface{} {
	return map[string]interface{}{"patternId": f.PatternId}
}

// GetUpdate returns the update data for this pattern record
func (f *FacilityAvailabilityPattern) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"patternType":       f.IsRepeatable,
		"weeklyPattern":     f.WeeklyPattern,
		"specificDates":     f.SpecificDates,
		"totalSpotsPerSlot": f.TotalSpotsPerSlot,
		"validFrom":         f.ValidFrom,
		"validUntil":        f.ValidUntil,
		"isActive":          f.IsActive,
		"updatedBy":         f.UpdatedBy,
		"updatedAt":         time.Now().Unix(),
	}
}

// MapData maps the provided data to the FacilityAvailabilityPattern struct
func (f *FacilityAvailabilityPattern) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, f)
}

// EncodeRedisData encodes the FacilityAvailabilityPattern data for Redis storage
func (f *FacilityAvailabilityPattern) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(f)
	return bytes
}

// DecodeRedisData decodes the FacilityAvailabilityPattern data from Redis storage
func (f *FacilityAvailabilityPattern) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, f)
}
