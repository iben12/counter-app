package db

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ParseFrequency parses a frequency string (e.g., "1h", "2d", "3w") and returns
// the number of units and the unit type (h, d, w).
func ParseFrequency(freq string) (int, string, error) {
	re := regexp.MustCompile(`^(\d+)([hdw])$`)
	matches := re.FindStringSubmatch(freq)
	if matches == nil {
		return 0, "", fmt.Errorf("invalid frequency format: %s (expected format: Nh, Nd, or Nw)", freq)
	}
	n, _ := strconv.Atoi(matches[1])
	unit := matches[2]
	return n, unit, nil
}

// NextExpiryTime calculates the next calendar-aligned expiry time based on the frequency and timezone.
// For "1h": top of the next hour (in the given timezone)
// For "2h": next 2-hour boundary from midnight (in the given timezone)
// For "1d": midnight of the next day (in the given timezone)
// For "2d": next 2-day boundary from epoch (in the given timezone)
// For "1w": midnight of the next Sunday (in the given timezone)
// For "2w": next 2-week boundary from epoch (in the given timezone)
//
// The calculation is performed in the given timezone, then converted back to UTC.
func NextExpiryTime(freq string, now time.Time, timezone string) (time.Time, error) {
	n, unit, err := ParseFrequency(freq)
	if err != nil {
		return time.Time{}, err
	}

	// Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", timezone)
	}

	// Convert to the counter's timezone for calculation
	nowInTZ := now.In(loc)

	switch unit {
	case "h":
		// Next N-hour boundary from midnight (in timezone)
		hoursSinceMidnight := nowInTZ.Hour()
		nextBoundaryHour := ((hoursSinceMidnight / n) + 1) * n
		if nextBoundaryHour >= 24 {
			// Roll over to next day at midnight
			midnightTomorrow := nowInTZ.AddDate(0, 0, 1).Truncate(24 * time.Hour)
			return midnightTomorrow, nil
		}
		midnightToday := nowInTZ.Truncate(24 * time.Hour)
		boundaryTZ := midnightToday.Add(time.Duration(nextBoundaryHour) * time.Hour)
		return boundaryTZ.UTC(), nil

	case "d":
		// Next N-day boundary from midnight of day 0 (Unix epoch Jan 1, 1970) in timezone
		epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, loc)
		daysSinceEpoch := int(nowInTZ.Sub(epoch).Hours() / 24)
		nextBoundaryDay := ((daysSinceEpoch / n) + 1) * n
		boundaryTZ := epoch.AddDate(0, 0, nextBoundaryDay).Truncate(24 * time.Hour)
		return boundaryTZ.UTC(), nil

	case "w":
		// Next N-week boundary: Sunday midnight N weeks from epoch (in timezone)
		// Epoch Jan 1, 1970 was a Thursday, so first Monday is Jan 4, 1970
		epoch := time.Date(1970, 1, 5, 0, 0, 0, 0, loc) // First Monday
		weeksSinceEpoch := int(nowInTZ.Sub(epoch).Hours() / (24 * 7))
		nextBoundaryWeek := ((weeksSinceEpoch / n) + 1) * n
		boundaryTZ := epoch.AddDate(0, 0, nextBoundaryWeek*7).Truncate(24 * time.Hour)
		return boundaryTZ.UTC(), nil

	default:
		return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
	}
}
