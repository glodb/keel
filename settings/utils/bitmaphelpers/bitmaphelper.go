package bitmaphelpers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
	"github.com/glodb/keel/settings/cachesettings/cache"
	"github.com/glodb/keel/settings/logger"
)

// BitmapHelper provides Redis bitmap operations for facility availability
type BitmapHelper struct{}

var getInstance = sync.OnceValue(func() *BitmapHelper {
	instance := &BitmapHelper{}
	return instance
})

// Singleton. Returns a single object of Utils
func GetBitmapHelper() *BitmapHelper {
	return getInstance()
}

// Constants for bitmap operations
const (
	MinutesPerDay        = 1440 // 24 hours * 60 minutes
	BitmapCacheTTL       = 7200 // 2 hours in seconds
	AvailabilityCacheTTL = 1800 // 30 minutes in seconds
)

// GenerateFacilityDateKey generates Redis key for facility-date bitmap
func (b *BitmapHelper) GenerateFacilityDateKey(facilityId string, date string) string {
	return fmt.Sprintf("facility:%s:availability:%s", facilityId, date)
}

// GenerateBookingBitmapKey generates Redis key for booking bitmap
func (b *BitmapHelper) GenerateBookingBitmapKey(facilityId string, date string) string {
	return fmt.Sprintf("facility:%s:bookings:%s", facilityId, date)
}

// TimeToMinute converts HH:MM format to minute of day (0-1439)
func (b *BitmapHelper) TimeToMinute(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return hours*60 + minutes, nil
}

// MinuteToTime converts minute of day to HH:MM format
func (b *BitmapHelper) MinuteToTime(minute int) string {
	hours := minute / 60
	mins := minute % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

// ParseDate parses YYYY-MM-DD date string to time.Time
func (b *BitmapHelper) ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// GetDayOfWeek returns day of week (0=Sunday, 1=Monday, ..., 6=Saturday)
func (b *BitmapHelper) GetDayOfWeek(dateStr string) (int, error) {
	t, err := b.ParseDate(dateStr)
	if err != nil {
		return 0, err
	}
	return int(t.Weekday()), nil
}

// GetDayName returns day name from date
func (b *BitmapHelper) GetDayName(dateStr string) (string, error) {
	t, err := b.ParseDate(dateStr)
	if err != nil {
		return "", err
	}
	return strings.ToLower(t.Weekday().String()), nil
}

// CreateAvailabilityBitmap creates a bitmap representing facility availability for a date
// Returns a byte array where each bit represents a minute (1 = available, 0 = not available)
func (b *BitmapHelper) CreateAvailabilityBitmap(pattern *keelmodels.FacilityAvailabilityPattern, dateStr string) ([]byte, error) {
	// Initialize bitmap with all zeros (not available)
	bitmap := make([]byte, (MinutesPerDay+7)/8) // 180 bytes for 1440 bits

	if pattern == nil || !pattern.IsActive {
		return bitmap, nil
	}

	var timeSlots []keelmodels.TimeSlot

	// Get time slots based on pattern type
	if pattern.IsRepeatable && pattern.WeeklyPattern != nil {
		// Get slots for the specific day of week
		dayName, err := b.GetDayName(dateStr)
		if err != nil {
			return bitmap, err
		}

		timeSlots = b.GetTimeSlotsForDay(pattern.WeeklyPattern, dayName)
	} else if !pattern.IsRepeatable && pattern.SpecificDates != nil {
		// Find the specific date in the list
		for _, specificDate := range pattern.SpecificDates {
			if specificDate.Date == dateStr {
				timeSlots = specificDate.TimeSlots
				break
			}
		}
	}

	// Set bits for available time slots
	for _, slot := range timeSlots {
		startMinute, err := b.TimeToMinute(slot.StartTime)
		if err != nil {
			logger.Log().Warn("Invalid start time", logger.StringField("time", slot.StartTime))
			continue
		}

		endMinute, err := b.TimeToMinute(slot.EndTime)
		if err != nil {
			logger.Log().Warn("Invalid end time", logger.StringField("time", slot.EndTime))
			continue
		}

		// Set bits from startMinute to endMinute-1
		for minute := startMinute; minute < endMinute && minute < MinutesPerDay; minute++ {
			byteIndex := minute / 8
			bitIndex := uint(minute % 8)
			bitmap[byteIndex] |= (1 << bitIndex)
		}
	}

	return bitmap, nil
}

// GetTimeSlotsForDay extracts time slots for a specific day from weekly schedule
func (b *BitmapHelper) GetTimeSlotsForDay(schedule *keelmodels.WeeklySchedule, dayName string) []keelmodels.TimeSlot {
	switch dayName {
	case "monday":
		return schedule.Monday
	case "tuesday":
		return schedule.Tuesday
	case "wednesday":
		return schedule.Wednesday
	case "thursday":
		return schedule.Thursday
	case "friday":
		return schedule.Friday
	case "saturday":
		return schedule.Saturday
	case "sunday":
		return schedule.Sunday
	default:
		return []keelmodels.TimeSlot{}
	}
}

// SetAvailabilityBitmapInRedis stores the availability bitmap in Redis with expiration
func (b *BitmapHelper) SetAvailabilityBitmapInRedis(facilityId string, dateStr string, bitmap []byte) error {
	key := b.GenerateFacilityDateKey(facilityId, dateStr)
	ctx := cache.GetCacheContext()

	// Store with expiration using SetEx
	err := cache.GetCache().SetEx(ctx, key, bitmap, BitmapCacheTTL)
	if err != nil {
		return err
	}

	return nil
}

// GetAvailabilityBitmapFromRedis retrieves the availability bitmap from Redis
func (b *BitmapHelper) GetAvailabilityBitmapFromRedis(facilityId string, dateStr string) ([]byte, error) {
	key := b.GenerateFacilityDateKey(facilityId, dateStr)
	ctx := cache.GetCacheContext()

	data, err := cache.GetCache().Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("bitmap not found")
	}

	return data, nil
}

// CreateBookingBitmap creates a bitmap representing all bookings for a facility on a date
// For each minute, we store the count of spots booked (not binary)
func (b *BitmapHelper) CreateBookingBitmap(bookings []keelmodels.FacilityBooking, dateStr string) map[int]int {
	// Map: minute -> spots booked
	spotsPerMinute := make(map[int]int)

	// Get day name for recurring bookings
	dayName, err := b.GetDayName(dateStr)
	if err != nil {
		return spotsPerMinute
	}

	for _, booking := range bookings {
		if booking.Status != "ACTIVE" {
			continue
		}

		var timeSlots []keelmodels.TimeSlot

		// Extract time slots based on booking pattern
		if booking.IsRepeatable && booking.WeeklyPattern != nil {
			// Recurring booking - check if it applies to this day
			timeSlots = b.GetTimeSlotsForDay(booking.WeeklyPattern, dayName)
		} else if !booking.IsRepeatable && len(booking.SpecificDates) > 0 {
			// Specific dates booking - check if it includes this date
			for _, dailySchedule := range booking.SpecificDates {
				if dailySchedule.Date == dateStr {
					timeSlots = append(timeSlots, dailySchedule.TimeSlots...)
				}
			}
		}

		// Add spots for each time slot
		for _, slot := range timeSlots {
			startMinute, err := b.TimeToMinute(slot.StartTime)
			if err != nil {
				continue
			}

			endMinute, err := b.TimeToMinute(slot.EndTime)
			if err != nil {
				continue
			}

			for minute := startMinute; minute < endMinute && minute < MinutesPerDay; minute++ {
				spotsPerMinute[minute] += booking.SpotsBooked
			}
		}
	}

	return spotsPerMinute
}

// CheckTimeSlotAvailability checks if a time slot is available using bitmaps
// Returns true if the slot is available with enough spots
func (b *BitmapHelper) CheckTimeSlotAvailability(
	availabilityBitmap []byte,
	bookingsMap map[int]int,
	startTime string,
	endTime string,
	totalSpots int,
	spotsNeeded int,
) bool {
	startMinute, err := b.TimeToMinute(startTime)
	if err != nil {
		return false
	}

	endMinute, err := b.TimeToMinute(endTime)
	if err != nil {
		return false
	}

	// Check each minute in the slot
	for minute := startMinute; minute < endMinute && minute < MinutesPerDay; minute++ {
		// Check if minute is in facility's available hours (bitmap check)
		byteIndex := minute / 8
		bitIndex := uint(minute % 8)

		if byteIndex >= len(availabilityBitmap) {
			return false
		}

		isAvailable := (availabilityBitmap[byteIndex] & (1 << bitIndex)) != 0
		if !isAvailable {
			return false // Facility not open at this minute
		}

		// Check if enough spots available
		bookedSpots := bookingsMap[minute]
		if bookedSpots+spotsNeeded > totalSpots {
			return false // Not enough spots
		}
	}

	return true
}

// GetOrCreateAvailabilityBitmap gets bitmap from cache or creates it
func (b *BitmapHelper) GetOrCreateAvailabilityBitmap(
	facilityId string,
	dateStr string,
	pattern *keelmodels.FacilityAvailabilityPattern,
) ([]byte, error) {
	// Try to get from cache first
	bitmap, err := b.GetAvailabilityBitmapFromRedis(facilityId, dateStr)
	if err == nil && bitmap != nil {
		return bitmap, nil
	}

	// Not in cache - create new bitmap
	bitmap, err = b.CreateAvailabilityBitmap(pattern, dateStr)
	if err != nil {
		return nil, err
	}

	// Store in cache
	err = b.SetAvailabilityBitmapInRedis(facilityId, dateStr, bitmap)
	if err != nil {
		logger.Log().Warn("Failed to cache availability bitmap", logger.StringField("error", err.Error()))
	}

	return bitmap, nil
}

// CalculateAvailableMinutes calculates how many minutes are available in a time range
func (b *BitmapHelper) CalculateAvailableMinutes(
	availabilityBitmap []byte,
	bookingsMap map[int]int,
	startTime string,
	endTime string,
	totalSpots int,
	spotsNeeded int,
) int {
	startMinute, err := b.TimeToMinute(startTime)
	if err != nil {
		return 0
	}

	endMinute, err := b.TimeToMinute(endTime)
	if err != nil {
		return 0
	}

	availableMinutes := 0

	for minute := startMinute; minute < endMinute && minute < MinutesPerDay; minute++ {
		byteIndex := minute / 8
		bitIndex := uint(minute % 8)

		if byteIndex >= len(availabilityBitmap) {
			continue
		}

		isAvailable := (availabilityBitmap[byteIndex] & (1 << bitIndex)) != 0
		if !isAvailable {
			continue
		}

		bookedSpots := bookingsMap[minute]
		if bookedSpots+spotsNeeded <= totalSpots {
			availableMinutes++
		}
	}

	return availableMinutes
}

// InvalidateCache invalidates the cache for a facility on a specific date
func (b *BitmapHelper) InvalidateCache(facilityId string, dateStr string) {
	ctx := cache.GetCacheContext()
	availKey := b.GenerateFacilityDateKey(facilityId, dateStr)
	bookingKey := b.GenerateBookingBitmapKey(facilityId, dateStr)

	cache.GetCache().Del(ctx, availKey)
	cache.GetCache().Del(ctx, bookingKey)
}

// InvalidateDateRange invalidates cache for a date range (useful for recurring bookings)
func (b *BitmapHelper) InvalidateDateRange(facilityId string, startDate string, endDate string) {
	start, err := b.ParseDate(startDate)
	if err != nil {
		return
	}

	end, err := b.ParseDate(endDate)
	if err != nil {
		return
	}

	ctx := context.Background()
	// Invalidate each date in range
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		b.InvalidateCache(facilityId, dateStr)
	}

	logger.Log().Info("Invalidated cache for date range",
		logger.StringField("facilityId", facilityId),
		logger.StringField("startDate", startDate),
		logger.StringField("endDate", endDate))

	_ = ctx // avoid unused warning
}
