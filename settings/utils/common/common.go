package common

import (
	"fmt"
	"sync"
	"time"

	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
	"github.com/glodb/keel/app/models/jsonmodels"
	"github.com/glodb/keel/settings/cachesettings/cache"
	"github.com/glodb/keel/settings/logger"

	"github.com/bytedance/sonic"
)

type Common struct {
}

var getInstance = sync.OnceValue(func() *Common {
	instance := &Common{}
	return instance
})

// Singleton. Returns a single object of Utils
func GetInstance() *Common {
	return getInstance()
}

func (u *Common) GetTimeSlots(isRepeatable bool, DailySchedules []keelmodels.DailySchedule, WeeklySchedule *keelmodels.WeeklySchedule, days int, startTime int64, addDay bool) []jsonmodels.TimeSlotWithUnix {
	var timeSlots []jsonmodels.TimeSlotWithUnix

	if isRepeatable {
		// Generate time slots for the next 3 months based on weekly schedule
		if WeeklySchedule == nil {
			return timeSlots
		}

		now := time.Unix(startTime, 0)
		if addDay {
			now = now.AddDate(0, 0, 1) // Add 1 day to the start time
		}
		endDate := now.AddDate(0, 0, days) // 3 months from now

		// Map day names to time.Weekday
		dayMap := map[string]time.Weekday{
			"monday":    time.Monday,
			"tuesday":   time.Tuesday,
			"wednesday": time.Wednesday,
			"thursday":  time.Thursday,
			"friday":    time.Friday,
			"saturday":  time.Saturday,
			"sunday":    time.Sunday,
		}

		// Map day names to their time slots
		scheduleMap := map[string][]keelmodels.TimeSlot{
			"monday":    WeeklySchedule.Monday,
			"tuesday":   WeeklySchedule.Tuesday,
			"wednesday": WeeklySchedule.Wednesday,
			"thursday":  WeeklySchedule.Thursday,
			"friday":    WeeklySchedule.Friday,
			"saturday":  WeeklySchedule.Saturday,
			"sunday":    WeeklySchedule.Sunday,
		}

		// Iterate through each day from now to 3 months ahead
		currentDate := now
		for currentDate.Before(endDate) {
			weekday := currentDate.Weekday()

			// Find the day name for this weekday
			var dayName string
			for name, day := range dayMap {
				if day == weekday {
					dayName = name
					break
				}
			}

			// Get time slots for this day
			if daySlots, exists := scheduleMap[dayName]; exists {
				for _, slot := range daySlots {
					// Parse time strings (format: "HH:MM")
					parsedStartTime, err := time.Parse("15:04", slot.StartTime)
					if err != nil {
						logger.Log().Error("Error parsing start time", logger.StringField("error", err.Error()), logger.StringField("time", slot.StartTime))
						continue
					}

					parsedEndTime, err := time.Parse("15:04", slot.EndTime)
					if err != nil {
						logger.Log().Error("Error parsing end time", logger.StringField("error", err.Error()), logger.StringField("time", slot.EndTime))
						continue
					}

					// Combine date with time
					startDateTime := time.Date(
						currentDate.Year(),
						currentDate.Month(),
						currentDate.Day(),
						parsedStartTime.Hour(),
						parsedStartTime.Minute(),
						0,
						0,
						time.UTC,
					)

					endDateTime := time.Date(
						currentDate.Year(),
						currentDate.Month(),
						currentDate.Day(),
						parsedEndTime.Hour(),
						parsedEndTime.Minute(),
						0,
						0,
						time.UTC,
					)
					// If the end time is before (or equal to) the start time, treat it as crossing midnight.
					if !endDateTime.After(startDateTime) {
						endDateTime = endDateTime.Add(24 * time.Hour)
					}

					// Only include slots that are in the future (after startTime parameter)
					if startDateTime.Unix() >= startTime {
						timeSlots = append(timeSlots, jsonmodels.TimeSlotWithUnix{
							StartTimeUnix: startDateTime.Unix(),
							EndTimeUnix:   endDateTime.Unix(),
						})
					}
				}
			}

			// Move to next day
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	} else {
		// Convert daily schedules to unix timestamps
		for _, dailySchedule := range DailySchedules {
			// Parse date (format: "DD-MM-YYYY" based on user's example "10-09-2025")
			date, err := time.Parse("02-01-2006", dailySchedule.Date)
			if err != nil {
				// Try alternative format "YYYY-MM-DD" (from struct comment)
				date, err = time.Parse("2006-01-02", dailySchedule.Date)
				if err != nil {
					logger.Log().Error("Error parsing date", logger.StringField("error", err.Error()), logger.StringField("date", dailySchedule.Date))
					continue
				}
			}

			// Process each time slot for this date
			for _, slot := range dailySchedule.TimeSlots {
				// Parse time strings (format: "HH:MM")
				parsedStartTime, err := time.Parse("15:04", slot.StartTime)
				if err != nil {
					logger.Log().Error("Error parsing start time", logger.StringField("error", err.Error()), logger.StringField("time", slot.StartTime))
					continue
				}

				parsedEndTime, err := time.Parse("15:04", slot.EndTime)
				if err != nil {
					logger.Log().Error("Error parsing end time", logger.StringField("error", err.Error()), logger.StringField("time", slot.EndTime))
					continue
				}

				// Combine date with time
				startDateTime := time.Date(
					date.Year(),
					date.Month(),
					date.Day(),
					parsedStartTime.Hour(),
					parsedStartTime.Minute(),
					0,
					0,
					time.UTC,
				)

				endDateTime := time.Date(
					date.Year(),
					date.Month(),
					date.Day(),
					parsedEndTime.Hour(),
					parsedEndTime.Minute(),
					0,
					0,
					time.UTC,
				)
				// If the end time is before (or equal to) the start time, treat it as crossing midnight.
				if !endDateTime.After(startDateTime) {
					endDateTime = endDateTime.Add(24 * time.Hour)
				}

				// Only include slots that are in the future (after startTime parameter)
				if startDateTime.Unix() >= startTime {
					timeSlots = append(timeSlots, jsonmodels.TimeSlotWithUnix{
						StartTimeUnix: startDateTime.Unix(),
						EndTimeUnix:   endDateTime.Unix(),
					})
				}
			}
		}
	}

	return timeSlots
}

func (u *Common) GetFacilityName(facilityId string) string {
	if facilityId == "meeting_point" {
		return "Meeting Point"
	}
	cacheData := cache.GetCache().HashGet(cache.GetCacheContext(), "facilities_id_map", facilityId)
	if len(cacheData) == 0 {
		return ""
	}
	var facility keelmodels.Facility
	err := sonic.Unmarshal([]byte(cacheData), &facility)
	if err != nil {
		return ""
	}
	return facility.Name
}

func (u *Common) GetFacilitiesNames(facilityIds []interface{}) (map[string]keelmodels.Facility, error) {

	facilityNames := map[string]keelmodels.Facility{}
	filteredFacilityIds := []interface{}{}
	for _, facilityId := range facilityIds {
		if facilityId == "meeting_point" {
			facilityNames[facilityId.(string)] = keelmodels.Facility{
				FacilityId: facilityId.(string),
				Name:       "Meeting Point",
			}
			continue
		}
		filteredFacilityIds = append(filteredFacilityIds, facilityId)
	}

	if len(filteredFacilityIds) <= 0 {
		return facilityNames, nil
	}

	cacheData := cache.GetCache().HashMGet(cache.GetCacheContext(), "facilities_id_map", filteredFacilityIds)
	if len(cacheData) == 0 {
		return facilityNames, nil // Return what we have so far (meeting_point if any)
	}

	for _, data := range cacheData {
		// Skip empty cache entries
		if len(data) == 0 {
			continue
		}

		var facility keelmodels.Facility
		err := sonic.Unmarshal([]byte(data), &facility)
		if err != nil {
			logger.Log().Warn("Failed to unmarshal facility, skipping",
				logger.StringField("data", data),
				logger.ErrorField("error", err))
			continue // Skip this facility but continue with others
		}

		// Add to map (this was missing!)
		facilityNames[facility.FacilityId] = facility
	}
	return facilityNames, nil
}

func (u *Common) GetVenuesNames(venueIds []interface{}) (map[string]keelmodels.Venue, error) {
	venueNames := map[string]keelmodels.Venue{}

	if len(venueIds) == 0 {
		return venueNames, nil
	}

	cacheData := cache.GetCache().HashMGet(cache.GetCacheContext(), "venues_id_map", venueIds)
	if len(cacheData) == 0 {
		return venueNames, nil // Return empty map instead of error
	}

	for _, data := range cacheData {
		// Skip empty cache entries
		if len(data) == 0 {
			continue
		}

		var venue keelmodels.Venue
		err := sonic.Unmarshal([]byte(data), &venue)
		if err != nil {
			logger.Log().Warn("Failed to unmarshal venue, skipping",
				logger.StringField("data", data),
				logger.ErrorField("error", err))
			continue // Skip this venue but continue with others
		}
		venueNames[venue.VenueId] = venue
	}
	return venueNames, nil
}

func (u *Common) GetThirtyMinuteSlots(startTime int64, endTime int64) []jsonmodels.TimeSlotWithUnix {
	var timeSlots []jsonmodels.TimeSlotWithUnix
	for startTime < endTime {
		timeSlots = append(timeSlots, jsonmodels.TimeSlotWithUnix{
			StartTimeUnix: startTime,
			EndTimeUnix:   startTime + 30*60,
		})
		startTime = startTime + 30*60
	}
	return timeSlots
}

func (u *Common) versionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			return ""
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

func (u *Common) CompareVersions(savedVersion, passedVersion string) bool {
	if savedVersion == "" {
		return true
	}

	latestVersion, appVersion := u.versionOrdinal(savedVersion), u.versionOrdinal(passedVersion)
	return appVersion >= latestVersion
}

func (u *Common) GetUserInfoFromCache(uiVal any, ok bool) (keelmodels.UserInfo, error) {
	var userInfo keelmodels.UserInfo
	if ui, ok2 := uiVal.(keelmodels.UserInfo); ok2 {
		userInfo = ui
	} else {
		// responses.SetResponse(c, http.StatusUnauthorized, responses.USER_NOT_LOGGED_IN, "unable to determine user from context", nil)
		return userInfo, fmt.Errorf("unable to determine user from context")
	}

	return userInfo, nil
}

// GetUserInfoFromCache gets user info from cache
func (cs *Common) GetUserInfoFromUserId(userId string) (*keelmodels.UserInfo, error) {
	userData, err := cache.GetCache().Get(cache.GetCacheContext(), fmt.Sprintf("%s_user_info", userId))
	if err != nil {
		return nil, err
	}
	if len(userData) == 0 {
		return nil, fmt.Errorf("user info not found in cache")
	}

	var userInfo keelmodels.UserInfo
	userInfo.DecodeRedisData(userData)
	return &userInfo, nil
}

func (u *Common) UpdateUserInfo(userInfo *keelmodels.UserInfo) error {
	cache.GetCache().Set(cache.GetCacheContext(), fmt.Sprintf("%s_user_info", userInfo.UserId), userInfo.EncodeRedisData())
	return nil
}

func (u *Common) GetUserSport(userSports []keelmodels.UserSports, sportId string) (*keelmodels.UserSports, error) {
	for _, userSport := range userSports {
		if userSport.SportId == sportId {
			return &userSport, nil
		}
	}
	return nil, fmt.Errorf("sport not found")
}
