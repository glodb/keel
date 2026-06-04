package keelmodels

import (
	"github.com/bytedance/sonic"
)

// TimeSlot represents a time slot in a schedule
type TimeSlot struct {
	StartTime string `json:"startTime" bson:"start_time"`
	EndTime   string `json:"endTime" bson:"end_time"`
}

// WeeklySchedule represents a user's weekly availability schedule for a specific sport
type WeeklySchedule struct {
	Monday    []TimeSlot `json:"monday" bson:"monday"`
	Tuesday   []TimeSlot `json:"tuesday" bson:"tuesday"`
	Wednesday []TimeSlot `json:"wednesday" bson:"wednesday"`
	Thursday  []TimeSlot `json:"thursday" bson:"thursday"`
	Friday    []TimeSlot `json:"friday" bson:"friday"`
	Saturday  []TimeSlot `json:"saturday" bson:"saturday"`
	Sunday    []TimeSlot `json:"sunday" bson:"sunday"`
	Timezone  string     `json:"timezone" bson:"timezone"`
	UpdatedAt int64      `json:"updatedAt" bson:"updatedAt"`
}

// DailySchedule represents a specific day's schedule with generated time slots
type DailySchedule struct {
	Date      string     `json:"date" bson:"date"` // Format: YYYY-MM-DD
	TimeSlots []TimeSlot `json:"timeSlots" bson:"timeSlots"`
	IsActive  bool       `json:"isActive" bson:"isActive"`
	CreatedAt int64      `json:"createdAt" bson:"createdAt"`
}

type SportSettings struct {
	SportId string `json:"sportId" bson:"sportId"`
	LevelId string `json:"levelId" bson:"levelId"`
}

type EventFacilityInfo struct {
	FacilityId string          `json:"facilityId" validate:"required"`
	BookingId  string          `json:"bookingId,omitempty"`
	Schedule   *WeeklySchedule `json:"schedule,omitempty"`
	Dates      []DailySchedule `json:"dates,omitempty"`
}

// UserSettings represents a user's settings for a specific sport
// Each sport will be a separate document in the userSettings collection
type UserSettings struct {
	UserId        string          `json:"userId" bson:"userId"`
	SportSettings []SportSettings `json:"sportSettings" bson:"sportSettings"`
	// Location fields - flat for easy indexing
	Latitude  float64 `json:"lat" bson:"lat"`
	Longitude float64 `json:"lng" bson:"lng"`
	Address   string  `json:"address,omitempty" bson:"address,omitempty"`
	City      string  `json:"city,omitempty" bson:"city,omitempty"`
	Country   string  `json:"country,omitempty" bson:"country,omitempty"`
	// Schedule
	Schedule        *WeeklySchedule `json:"schedule" bson:"schedule"`
	DailySchedules  []DailySchedule `json:"dailySchedules" bson:"dailySchedules"`
	IsRepeatable    bool            `json:"isRepeatable" bson:"isRepeatable"`
	IsPrivate       bool            `json:"isPrivate" bson:"isPrivate"`
	IsActive        bool            `json:"isActive" bson:"isActive"`
	CreatedAt       int64           `json:"createdAt" bson:"createdAt"`
	UpdatedAt       int64           `json:"updatedAt" bson:"updatedAt"`
	LastEndDate     int64           `json:"lastEndDate" bson:"lastEndDate"`
	LastProcessedAt int64           `json:"lastProcessedAt" bson:"lastProcessedAt"`
}

func (u *UserSettings) String() string {
	json, _ := sonic.Marshal(u)
	return string(json)
}
