package jsonmodels

import (
	"time"

	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
)

// TimeSlotWithUnix represents a time slot with unix timestamps
type TimeSlotWithUnix struct {
	StartTimeUnix int64
	EndTimeUnix   int64
}

// SearchPalsRequest represents the request for searching pals
type SearchPalsRequest struct {
	Type           string                     `json:"type,omitempty"`
	StartTime      int64                      `json:"startTime" validate:"required"`
	EndTime        int64                      `json:"endTime" validate:"required"`
	Timezone       string                     `json:"timezone,omitempty"`       // IANA timezone (e.g., "America/New_York") or empty for UTC
	TimezoneOffset int                        `json:"timezoneOffset,omitempty"` // Offset in minutes from UTC (e.g., -300 for EST)
	UseFilters     bool                       `json:"useFilters,omitempty"`
	VenueId        string                     `json:"venueId,omitempty"`
	IsRepeatable   bool                       `json:"isRepeatable,omitempty"`
	DailySchedules []keelmodels.DailySchedule `json:"dailySchedules,omitempty"`
	Schedule       keelmodels.WeeklySchedule  `json:"schedule,omitempty"`
	UserSports     []keelmodels.UserSports    `json:"userSports,omitempty"`
	Lat            float64                    `json:"lat,omitempty"`
	Lng            float64                    `json:"lng,omitempty"`
	City           string                     `json:"city,omitempty"`
	SearchQuery    string                     `json:"searchQuery,omitempty"`
	Radius         int                        `json:"radius,omitempty"`
	Limit          int                        `json:"limit,omitempty"`
	Page           int                        `json:"page,omitempty"`
	Sort           int                        `json:"sort,omitempty"`
	SortValue      string                     `json:"sortValue,omitempty"`
	RangeKey       string                     `json:"rangeKey,omitempty"`
}

// GetTimezone returns the timezone location, defaults to UTC if not specified or invalid
func (s *SearchPalsRequest) GetTimezone() *time.Location {
	if s.Timezone != "" {
		loc, err := time.LoadLocation(s.Timezone)
		if err == nil {
			return loc
		}
	}
	// Fall back to offset if timezone string is not provided
	if s.TimezoneOffset != 0 {
		return time.FixedZone("UserTZ", s.TimezoneOffset*60)
	}
	return time.UTC
}

func (s *SearchPalsRequest) GetPage() int {
	return s.Page
}
func (s *SearchPalsRequest) GetLimit() int {
	return s.Limit
}
func (s *SearchPalsRequest) GetSort() int {
	return s.Sort
}
func (s *SearchPalsRequest) GetSortValue() string {
	return s.SortValue
}
func (s *SearchPalsRequest) GetRangeKey() string {
	return s.RangeKey
}
func (s *SearchPalsRequest) GetStartTime() int64 {
	return s.StartTime
}
func (s *SearchPalsRequest) GetEndTime() int64 {
	return s.EndTime
}

func (s *SearchPalsRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	s.Page = page
	s.Limit = limit
	s.Sort = sort
	s.SortValue = sortValue
	s.RangeKey = rangeKey
	s.StartTime = startTime
	s.EndTime = endTime
}

// SearchPalsByVenueRequest represents the request for searching pals by venue
type SearchPalsByVenueRequest struct {
	VenueId   string `json:"venueId" validate:"required"`
	StartTime *int64 `json:"startTime,omitempty"` // Optional, defaults to now
	Limit     int    `json:"limit,omitempty"`
	Page      int    `json:"page,omitempty"`
	Sort      int    `json:"sort,omitempty"`
	SortValue string `json:"sortValue,omitempty"`
	RangeKey  string `json:"rangeKey,omitempty"`
}

func (s *SearchPalsByVenueRequest) GetPage() int {
	if s.Page < 1 {
		return 1
	}
	return s.Page
}

func (s *SearchPalsByVenueRequest) GetLimit() int {
	if s.Limit < 1 {
		return 20
	}
	return s.Limit
}

func (s *SearchPalsByVenueRequest) GetSort() int {
	return s.Sort
}

func (s *SearchPalsByVenueRequest) GetSortValue() string {
	return s.SortValue
}

func (s *SearchPalsByVenueRequest) GetRangeKey() string {
	return s.RangeKey
}

func (s *SearchPalsByVenueRequest) GetStartTime() int64 {
	if s.StartTime != nil {
		return *s.StartTime
	}
	return time.Now().Unix()
}

func (s *SearchPalsByVenueRequest) GetEndTime() int64 {
	return 0 // Not used for this query
}

func (s *SearchPalsByVenueRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	s.Page = page
	s.Limit = limit
	s.Sort = sort
	s.SortValue = sortValue
	s.RangeKey = rangeKey
	if startTime > 0 {
		s.StartTime = &startTime
	}
}

// UpdateEventSettingsRequest represents the request for updating event settings
type UpdateEventSettingsRequest struct {
	PalId                  string               `json:"palId" validate:"required"`
	TotalSpots             *int                 `json:"totalSpots,omitempty"`
	EnrollmentClosingHours *int                 `json:"enrollmentClosingHours,omitempty"` // Hours before startTime
	InvitationList         []keelmodels.PalInfo `json:"invitationList,omitempty"`
}

// ReportPalRequest is the body for reporting a user (palId) for moderation review.
type ReportPalRequest struct {
	PalId               string `json:"palId" validate:"required"`
	ReportedReason      string `json:"reportedReason,omitempty"`
	ReportedDescription string `json:"reportedDescription,omitempty"`
	ReportedStatus      string `json:"reportedStatus,omitempty"`
	ReportedType        string `json:"reportedType,omitempty"`
	ReportedSeverity    string `json:"reportedSeverity,omitempty"`
	ReportedCategory    string `json:"reportedCategory,omitempty"`
}

// ClearPalReportRequest clears an active report aggregate and deletes its detail rows.
type ClearPalReportRequest struct {
	ReportId      string `json:"reportId" validate:"required"`
	ClearedReason string `json:"clearedReason,omitempty"`
}

// SuggestLevelChangeRequest represents the request for suggesting a level change
type SuggestLevelChangeRequest struct {
	TargetUserId string `json:"targetUserId" validate:"required" field:"targetUserId"`
	SportId      string `json:"sportId" validate:"required" field:"sportId"`
	LevelId      string `json:"levelId" validate:"required" field:"levelId"`
}
