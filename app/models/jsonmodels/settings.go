package jsonmodels

import "github.com/glodb/keel/app/models/dbmodels/keelmodels"

// AddUserSettingsRequest models the accepted payload
// repeatable=true => weekly schedule must be provided in schedule
// repeatable=false => one-off dates must be provided in dates
type AddUserSettingsRequest struct {
	SportsId string `json:"sportsId" validate:"required"`
	LevelId  string `json:"levelId" validate:"required"`
	Location struct {
		Lat     float64 `json:"lat" validate:"required"`
		Lng     float64 `json:"lng" validate:"required"`
		Address string  `json:"address"`
		City    string  `json:"city"`
		Country string  `json:"country"`
	} `json:"location" validate:"required"`
	Repeatable bool                       `json:"repeatable"`
	Schedule   *keelmodels.WeeklySchedule `json:"schedule,omitempty"`
	Dates      []keelmodels.DailySchedule `json:"dates,omitempty"`
}

// New granular request models
type AddUserSportRequest struct {
	SportsId string `json:"sportsId" validate:"required"`
	LevelId  string `json:"levelId" validate:"required"`
}

// New granular request models
type RemoveUserSportRequest struct {
	SportsId string `json:"sportsId" validate:"required"`
}

type AddUserSportsRequest struct {
	Sports []AddUserSportRequest `json:"sports" validate:"required,min=1,dive"`
}

// UpdateUserSportsRequest uses the same structure as AddUserSportsRequest
// but will replace all existing sports with the new ones
type UpdateUserSportsRequest struct {
	Sports []AddUserSportRequest `json:"sports" validate:"required,min=1,dive"`
}

type AddUserLocationRequest struct {
	Lat     float64 `json:"lat" validate:"required"`
	Lng     float64 `json:"lng" validate:"required"`
	Address string  `json:"address"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	Radius  int     `json:"radius"`
}

type AddUserScheduleRequest struct {
	Repeatable bool                       `json:"repeatable"`
	Schedule   *keelmodels.WeeklySchedule `json:"schedule,omitempty"`
	Dates      []keelmodels.DailySchedule `json:"dates,omitempty"`
}

type SearchUserSportsRequest struct {
	SportsId    string `json:"sportsId" validate:"required"`
	SearchQuery string `json:"searchQuery"`
	Limit       int    `json:"limit,omitempty"`
	Page        int    `json:"page,omitempty"`
	Sort        int    `json:"sort,omitempty"`
	SortValue   string `json:"sortValue,omitempty"`
	RangeKey    string `json:"rangeKey,omitempty"`
	StartTime   int64  `json:"startTime,omitempty"`
	EndTime     int64  `json:"endTime,omitempty"`
}

func (s *SearchUserSportsRequest) GetPage() int {
	return s.Page
}
func (s *SearchUserSportsRequest) GetLimit() int {
	return s.Limit
}
func (s *SearchUserSportsRequest) GetSort() int {
	return s.Sort
}
func (s *SearchUserSportsRequest) GetSortValue() string {
	return s.SortValue
}
func (s *SearchUserSportsRequest) GetRangeKey() string {
	return s.RangeKey
}
func (s *SearchUserSportsRequest) GetStartTime() int64 {
	return s.StartTime
}
func (s *SearchUserSportsRequest) GetEndTime() int64 {
	return s.EndTime
}

func (s *SearchUserSportsRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	s.Page = page
	s.Limit = limit
	s.Sort = sort
	s.SortValue = sortValue
	s.RangeKey = rangeKey
	s.StartTime = startTime
	s.EndTime = endTime
}
