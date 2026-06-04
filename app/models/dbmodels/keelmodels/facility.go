package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// Facility represents a facility entity in the database
// Note: TimeSlots have been moved to FacilityAvailabilityPattern
type Facility struct {
	FacilityId  string   `json:"facilityId" bson:"facilityId"`
	PatternId   string   `json:"patternId" bson:"patternId"`
	VenueId     string   `json:"venueId" bson:"venueId"`   // Reference to venue
	OwnerId     string   `json:"ownerId" bson:"ownerId"`   // Reference to owner
	Managers    []string `json:"managers" bson:"managers"` // Array of manager userIds
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	Image       string   `json:"image,omitempty" bson:"image,omitempty"`
	Icon        string   `json:"icon,omitempty" bson:"icon,omitempty"`
	Price       float64  `json:"price" bson:"price"`
	SportId     string   `json:"sportId" bson:"sportId"`
	CreatedAt   int64    `json:"createdAt" bson:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt" bson:"updatedAt"`
	CreatedBy   string   `json:"createdBy" bson:"createdBy"`                     // UserId of creator
	UpdatedBy   string   `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"` // UserId of last updater
}

// GetQuery returns the query for finding this facility record
func (f *Facility) GetQuery() map[string]interface{} {
	return map[string]interface{}{"facilityId": f.FacilityId}
}

func (f *Facility) GetSortKeyData(key string) any {
	if key == "createdAt" {
		return f.CreatedAt
	}
	if key == "updatedAt" {
		return f.UpdatedAt
	}
	if key == "name" {
		return f.Name
	}
	if key == "description" {
		return f.Description
	}
	if key == "price" {
		return f.Price
	}
	if key == "sportId" {
		return f.SportId
	}
	return f.FacilityId
}

func (f *Facility) GetTimeValue(timeKey string) int64 {
	if timeKey == "createdAt" {
		return f.CreatedAt
	}
	if timeKey == "updatedAt" {
		return f.UpdatedAt
	}
	return 0
}

// GetUpdate returns the update data for this facility record
func (f *Facility) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"name":        f.Name,
		"description": f.Description,
		"image":       f.Image,
		"icon":        f.Icon,
		"price":       f.Price,
		"sportId":     f.SportId,
		"updatedBy":   f.UpdatedBy,
		"updatedAt":   time.Now().Unix(),
	}
}

// MapData maps the provided data to the Facility struct
func (f *Facility) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, f)
}

// EncodeRedisData encodes the Facility data for Redis storage
func (f *Facility) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(f)
	return bytes
}

// DecodeRedisData decodes the Facility data from Redis storage
func (f *Facility) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, f)
}
