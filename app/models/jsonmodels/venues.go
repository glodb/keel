package jsonmodels

import (
	"time"

	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
)

// CreateVenueRequest models the request to create a venue
type CreateVenueRequest struct {
	Name           string                     `json:"name" validate:"required"`
	Description    string                     `json:"description"`
	IsPrivate      bool                       `json:"isPrivate"`
	Images         []string                   `json:"images"`
	Lat            float64                    `json:"lat" validate:"required"`
	Lng            float64                    `json:"lng" validate:"required"`
	WeeklySchedule *keelmodels.WeeklySchedule `json:"weeklySchedule" validate:"required"`
	Address        string                     `json:"address"`
	City           string                     `json:"city"`
	Country        string                     `json:"country"`
}

// UpdateVenueRequest models the request to update a venue
type UpdateVenueRequest struct {
	VenueId        string                     `json:"venueId" validate:"required"`
	WeeklySchedule *keelmodels.WeeklySchedule `json:"weeklySchedule,omitempty"`
	Name           string                     `json:"name,omitempty"`
	Description    string                     `json:"description,omitempty"`
	IsPrivate      *bool                      `json:"isPrivate,omitempty"`
	Images         []string                   `json:"images,omitempty"`
	Lat            *float64                   `json:"lat,omitempty"`
	Lng            *float64                   `json:"lng,omitempty"`
	Address        string                     `json:"address,omitempty"`
	City           string                     `json:"city,omitempty"`
	Country        string                     `json:"country,omitempty"`
	Sports         []string                   `json:"sports,omitempty"`
}

// AddManagerRequest models the request to add a manager to a venue
type AddManagerRequest struct {
	VenueId string `json:"venueId" validate:"required"`
	UserId  string `json:"userId" validate:"required"` // UserId of the manager to add
}

// RemoveManagerRequest models the request to remove a manager from a venue
type RemoveManagerRequest struct {
	VenueId string `json:"venueId" validate:"required"`
	UserId  string `json:"userId" validate:"required"` // UserId of the manager to remove
}

// DeleteVenueRequest models the request to delete a venue
type DeleteVenueRequest struct {
	VenueId string `json:"venueId" validate:"required" uri:"venueId"`
}

// ReportVenueRequest models the request to report a venue
type ReportVenueRequest struct {
	VenueId             string `json:"venueId" validate:"required"`
	ReportedReason      string `json:"reportedReason,omitempty"`
	ReportedDescription string `json:"reportedDescription,omitempty"`
	ReportedStatus      string `json:"reportedStatus,omitempty"`
	ReportedType        string `json:"reportedType,omitempty"`
	ReportedSeverity    string `json:"reportedSeverity,omitempty"`
	ReportedCategory    string `json:"reportedCategory,omitempty"`
}

// ClearVenueReportRequest models the request to clear a venue report
type ClearVenueReportRequest struct {
	ReportId      string `json:"reportId" validate:"required"`
	ClearedReason string `json:"clearedReason,omitempty"`
}

// GetVenueReportsRequest models the request to get venue reports
type GetVenueReportsRequest struct {
	VenueId string `json:"venueId" validate:"required" uri:"venueId"`
}

// AddOrEditVenueReviewRequest models the request to add or edit a venue review
type AddOrEditVenueReviewRequest struct {
	VenueId string `json:"venueId" validate:"required"`
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required,min=1,max=500"`
}

// DeleteVenueReviewRequest models the request to delete a venue review
type DeleteVenueReviewRequest struct {
	VenueId string `json:"venueId" validate:"required"`
}

// GetVenueReviewsRequest models the request to get venue reviews (paginated)
type GetVenueReviewsRequest struct {
	VenueId   string `json:"venueId" validate:"required"`
	Limit     int    `json:"limit,omitempty"`
	Page      int    `json:"page,omitempty"`
	Sort      string `json:"sort,omitempty"`
	SortValue int    `json:"sortValue,omitempty"`
	RangeKey  string `json:"rangeKey,omitempty"`
}

// GetPage returns the page number (default 1)
func (r *GetVenueReviewsRequest) GetPage() int {
	if r.Page < 1 {
		return 1
	}
	return r.Page
}

// GetLimit returns the limit (default 20)
func (r *GetVenueReviewsRequest) GetLimit() int {
	if r.Limit < 1 {
		return 20
	}
	return r.Limit
}

// GetSort returns the sort field
func (r *GetVenueReviewsRequest) GetSort() int {
	return 1 // Default ascending
}

// GetSortValue returns the sort value as string
func (r *GetVenueReviewsRequest) GetSortValue() string {
	if r.SortValue == -1 {
		return "-1"
	}
	return "1"
}

// GetRangeKey returns the range key
func (r *GetVenueReviewsRequest) GetRangeKey() string {
	return r.RangeKey
}

// GetStartTime returns the start time (not used for reviews)
func (r *GetVenueReviewsRequest) GetStartTime() int64 {
	return 0
}

// GetEndTime returns the end time (not used for reviews)
func (r *GetVenueReviewsRequest) GetEndTime() int64 {
	return 0
}

// Set method to update pagination values
func (r *GetVenueReviewsRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	r.Page = page
	r.Limit = limit
	r.Sort = ""
	r.SortValue = sort
	r.RangeKey = rangeKey
}

// SearchVenuesRequest models the request to search venues by location
type SearchVenuesRequest struct {
	Lat            float64                    `json:"lat" validate:"required"`
	Lng            float64                    `json:"lng" validate:"required"`
	StartTime      int64                      `json:"startTime" validate:"required"`
	EndTime        int64                      `json:"endTime" validate:"required"`
	Timezone       string                     `json:"timezone,omitempty"`       // IANA timezone (e.g., "America/New_York") or empty for UTC
	TimezoneOffset int                        `json:"timezoneOffset,omitempty"` // Offset in minutes from UTC (e.g., -300 for EST)
	Radius         *float64                   `json:"radius,omitempty"`         // In kilometers, default 5km
	UseFilters     bool                       `json:"useFilters,omitempty"`
	IsRepeatable   bool                       `json:"isRepeatable,omitempty"`
	DailySchedules []keelmodels.DailySchedule `json:"dailySchedules,omitempty"`
	Schedule       keelmodels.WeeklySchedule  `json:"schedule,omitempty"`
	UserSports     []keelmodels.UserSports    `json:"userSports,omitempty"`
	City           string                     `json:"city,omitempty"`
	Name           string                     `json:"name,omitempty"`      // Venue name search (fuzzy)
	Sports         []string                   `json:"sports,omitempty"`    // Array of sport IDs   // Unix timestamp
	Limit          int                        `json:"limit,omitempty"`     // Default 10
	Page           int                        `json:"page,omitempty"`      // Default 1
	Sort           string                     `json:"sort,omitempty"`      // Sort field
	SortValue      int                        `json:"sortValue,omitempty"` // 1 for asc, -1 for desc
	RangeKey       string                     `json:"rangeKey,omitempty"`  // For range queries
}

// GetTimezone returns the timezone location, defaults to UTC if not specified or invalid
func (r *SearchVenuesRequest) GetTimezone() *time.Location {
	if r.Timezone != "" {
		loc, err := time.LoadLocation(r.Timezone)
		if err == nil {
			return loc
		}
	}
	// Fall back to offset if timezone string is not provided
	if r.TimezoneOffset != 0 {
		return time.FixedZone("UserTZ", r.TimezoneOffset*60)
	}
	return time.UTC
}

// GetPage returns the page number (default 1)
func (r *SearchVenuesRequest) GetPage() int {
	if r.Page < 1 {
		return 1
	}
	return r.Page
}

// GetLimit returns the limit (default 10)
func (r *SearchVenuesRequest) GetLimit() int {
	if r.Limit < 1 {
		return 10
	}
	return r.Limit
}

// GetSort returns the sort field (1 for asc, -1 for desc)
func (r *SearchVenuesRequest) GetSort() int {
	return 1 // Default ascending
}

// GetSortValue returns the sort value as string
func (r *SearchVenuesRequest) GetSortValue() string {
	if r.SortValue == -1 {
		return "-1"
	}
	return "1"
}

// GetRangeKey returns the range key
func (r *SearchVenuesRequest) GetRangeKey() string {
	return r.RangeKey
}

// GetStartTime returns the start time
func (r *SearchVenuesRequest) GetStartTime() int64 {
	return r.StartTime
}

// GetEndTime returns the end time
func (r *SearchVenuesRequest) GetEndTime() int64 {
	return r.EndTime
}

// Set method to update pagination values
func (r *SearchVenuesRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	r.Page = page
	r.Limit = limit
	r.Sort = ""
	r.SortValue = sort
	r.RangeKey = rangeKey
	r.StartTime = startTime
	r.EndTime = endTime
}

// GetRadius returns the radius in kilometers (default 5)
func (r *SearchVenuesRequest) GetRadius() float64 {
	if r.Radius != nil && *r.Radius > 0 {
		return *r.Radius
	}
	return 5.0 // Default 5km
}

// GetNearestVenueRequest models the request to get the nearest venue
type GetNearestVenueRequest struct {
	PalId  string   `json:"palId,omitempty"`  // Optional - if provided, calculates midpoint between current user and pal
	Lat    *float64 `json:"lat,omitempty"`    // Optional - uses user profile location if not provided
	Lng    *float64 `json:"lng,omitempty"`    // Optional - uses user profile location if not provided
	Sports []string `json:"sports,omitempty"` // Optional - uses user profile sports if not provided
	Radius *float64 `json:"radius,omitempty"` // In kilometers, default 50km for nearest search
}

// GetRadius returns the radius in kilometers (default 50km for nearest venue search)
func (r *GetNearestVenueRequest) GetRadius() float64 {
	if r.Radius != nil && *r.Radius > 0 {
		return *r.Radius
	}
	return 50.0 // Default 50km for nearest venue search
}
