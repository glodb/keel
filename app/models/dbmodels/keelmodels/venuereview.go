package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// VenueReview represents a review for a venue
type VenueReview struct {
	ReviewId  string `json:"reviewId" bson:"reviewId"`
	VenueId   string `json:"venueId" bson:"venueId"`
	UserId    string `json:"userId" bson:"userId"`
	Rating    int    `json:"rating" bson:"rating"`       // Rating from 1-5
	Comment   string `json:"comment" bson:"comment"`     // Review comment
	CreatedAt int64  `json:"createdAt" bson:"createdAt"` // Timestamp when review was created
	UpdatedAt int64  `json:"updatedAt" bson:"updatedAt"` // Timestamp when review was last updated
}

// GetQuery returns the query for finding this review record
func (r *VenueReview) GetQuery() map[string]interface{} {
	return map[string]interface{}{
		"venueId": r.VenueId,
		"userId":  r.UserId,
	}
}

// GetUpdate returns the update data for this review record
func (r *VenueReview) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"rating":    r.Rating,
		"comment":   r.Comment,
		"updatedAt": time.Now().Unix(),
	}
}

// MapData maps the provided data to the VenueReview struct
func (r *VenueReview) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, r)
}

// EncodeRedisData encodes the VenueReview data for Redis storage
func (r *VenueReview) EncodeRedisData() []byte {
	bytes, _ := sonic.Marshal(r)
	return bytes
}

// DecodeRedisData decodes the VenueReview data from Redis storage
func (r *VenueReview) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, r)
}
