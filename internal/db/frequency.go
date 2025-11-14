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

// NextExpiryTime calculates the next calendar-aligned expiry time based on the frequency.
// For "1h": top of the next hour
// For "2h": next 2-hour boundary from midnight
// For "1d": midnight of the next day
// For "2d": next 2-day boundary from epoch
// For "1w": midnight of the next Sunday
// For "2w": next 2-week boundary from epoch (Sunday)
func NextExpiryTime(freq string, now time.Time) (time.Time, error) {
	n, unit, err := ParseFrequency(freq)
	if err != nil {
		return time.Time{}, err
	}

	// Normalize to UTC for consistent calculations
	now = now.UTC()

	switch unit {
	case "h":
		// Next N-hour boundary from midnight
		// Calculate hours since midnight, then round up to next N-hour boundary
		hoursSinceMidnight := now.Hour()
		nextBoundaryHour := ((hoursSinceMidnight / n) + 1) * n
		if nextBoundaryHour >= 24 {
			// Roll over to next day at midnight
			return now.AddDate(0, 0, 1).UTC().Truncate(24 * time.Hour), nil
		}
		midnight := now.Truncate(24 * time.Hour)
		return midnight.Add(time.Duration(nextBoundaryHour) * time.Hour), nil

	case "d":
		// Next N-day boundary from midnight of day 0 (Unix epoch Jan 1, 1970)
		epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		daysSinceEpoch := int(now.Sub(epoch).Hours() / 24)
		nextBoundaryDay := ((daysSinceEpoch / n) + 1) * n
		boundaryTime := epoch.AddDate(0, 0, nextBoundaryDay)
		return boundaryTime, nil

	case "w":
		// Next N-week boundary: Sunday midnight N weeks from epoch
		// Epoch Jan 1, 1970 was a Thursday, so first Sunday is Jan 4, 1970
		epoch := time.Date(1970, 1, 4, 0, 0, 0, 0, time.UTC) // First Sunday
		weeksSinceEpoch := int(now.Sub(epoch).Hours() / (24 * 7))
		nextBoundaryWeek := ((weeksSinceEpoch / n) + 1) * n
		boundaryTime := epoch.AddDate(0, 0, nextBoundaryWeek*7)
		return boundaryTime, nil

	default:
		return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
	}
}
