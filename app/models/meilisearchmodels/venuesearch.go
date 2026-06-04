package meilisearchmodels

const VenuesIndexName = "venues"

// VenueSearchDocument represents the venue document structure for Meilisearch
// Note: Uses GeoPoint from palsearch.go for _geo field
type VenueSearchDocument struct {
	VenueId      string   `json:"venueId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Images       []string `json:"images"`
	IsPrivate    bool     `json:"isPrivate"`
	IsFree       bool     `json:"isFree"`
	Price        float64  `json:"price"`
	OwnerId      string   `json:"ownerId"`
	Managers     []string `json:"managers"`
	Sports       []string `json:"sports"`
	Geo          GeoPoint `json:"_geo"` // MeiliSearch geo field for _geoRadius filtering
	Lat          float64  `json:"lat"`  // Keep for backward compatibility
	Lng          float64  `json:"lng"`  // Keep for backward compatibility
	Address      string   `json:"address"`
	City         string   `json:"city"`
	Country      string   `json:"country"`
	TotalReviews int      `json:"totalReviews"`
	TotalRatings int      `json:"totalRatings"`
	ReportCount  int      `json:"reportCount"`
	CreatedAt    int64    `json:"createdAt"`
	UpdatedAt    int64    `json:"updatedAt"`
}
