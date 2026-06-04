package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// VenuePrice represents pricing information for a venue
type VenuePrice struct {
	IsFree bool    `json:"isFree" bson:"isFree"`
	Price  float64 `json:"price,omitempty" bson:"price,omitempty"` // Required if IsFree is false
}

// Venue represents a venue entity in the database
type Venue struct {
	VenueId        string          `json:"venueId" bson:"venueId"`
	Name           string          `json:"name" bson:"name"`
	Description    string          `json:"description" bson:"description"`
	Images         []string        `json:"images" bson:"images"`
	IsPrivate      bool            `json:"isPrivate" bson:"isPrivate"`
	Pricing        VenuePrice      `json:"pricing" bson:"pricing"`
	OwnerId        string          `json:"ownerId" bson:"ownerId"`   // UserId of the venue owner
	Managers       []string        `json:"managers" bson:"managers"` // Array of userIds (owner is always included)
	Reports        []VenueReport   `json:"reports" bson:"reports"`   // Array of reports
	Sports         []string        `json:"sports" bson:"sports"`     // Array of sports
	WeeklySchedule *WeeklySchedule `json:"weeklySchedule,omitempty" bson:"weeklySchedule,omitempty"`
	Lat            float64         `json:"lat" bson:"lat"`
	Lng            float64         `json:"lng" bson:"lng"`
	Distance       float64         `json:"distance,omitempty" bson:"-"` // Calculated distance from user (not stored in DB)
	Address        string          `json:"address" bson:"address"`
	City           string          `json:"city" bson:"city"`
	Availability   int             `json:"availability" bson:"availability"` // 0: not available, 1: availble for booking, 2: available for playing event
	Country        string          `json:"country" bson:"country"`
	TotalReviews   int             `json:"totalReviews" bson:"totalReviews"` // Total number of reviews
	TotalRatings   int             `json:"totalRatings" bson:"totalRatings"` // Total number of ratings
	ReportCount    int             `json:"reportCount" bson:"reportCount"`   // Total number of reports
	CreatedAt      int64           `json:"createdAt" bson:"createdAt"`
	UpdatedAt      int64           `json:"updatedAt" bson:"updatedAt"`
}

func (v *Venue) GetTimeValue(timeKey string) int64 {
	switch timeKey {
	case "createdAt":
		return v.CreatedAt
	case "updatedAt":
		return v.UpdatedAt
	default:
		return v.CreatedAt
	}
}

func (v *Venue) GetSortKeyData(key string) any {
	switch key {
	case "createdAt":
		return v.CreatedAt
	case "updatedAt":
		return v.UpdatedAt
	case "name":
		return v.Name
	case "description":
		return v.Description
	case "isPrivate":
		return v.IsPrivate
	}
	return v.VenueId
}

// GetQuery returns the query for finding this venue record
func (v *Venue) GetQuery() map[string]interface{} {
	return map[string]interface{}{"venueId": v.VenueId}
}

// GetUpdate returns the update data for this venue record
func (v *Venue) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"name":        v.Name,
		"description": v.Description,
		"isPrivate":   v.IsPrivate,
		"pricing":     v.Pricing,
		"images":      v.Images,
		"managers":    v.Managers,
		"reports":     v.Reports,
		"reportCount": v.ReportCount,
		"updatedAt":   time.Now().Unix(),
	}
}

// MapData maps the provided data to the Venue struct
func (v *Venue) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, v)
}

// EncodeRedisData encodes the Venue data for Redis storage
func (v *Venue) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(v)
	return bytes
}

// DecodeRedisData decodes the Venue data from Redis storage
func (v *Venue) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, v)
}
