package db

import (
	"fmt"
	"testing"
	"time"
)

func TestParseFrequency(t *testing.T) {
	tests := []struct {
		freq    string
		wantN   int
		wantU   string
		wantErr bool
	}{
		{"1h", 1, "h", false},
		{"2d", 2, "d", false},
		{"3w", 3, "w", false},
		{"10h", 10, "h", false},
		{"invalid", 0, "", true},
		{"1m", 0, "", true},
		{"1", 0, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.freq, func(t *testing.T) {
			n, u, err := ParseFrequency(tt.freq)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFrequency() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (n != tt.wantN || u != tt.wantU) {
				t.Errorf("ParseFrequency() = (%d, %s), want (%d, %s)", n, u, tt.wantN, tt.wantU)
			}
		})
	}
}

func TestNextExpiryTime(t *testing.T) {
	// Test cases with known times
	tests := []struct {
		name      string
		freq      string
		now       time.Time
		wantHour  int
		wantDay   int
		wantMonth time.Month
		wantYear  int
	}{
		{
			name:      "1h at 14:30",
			freq:      "1h",
			now:       time.Date(2025, 11, 14, 14, 30, 0, 0, time.UTC),
			wantHour:  15,
			wantDay:   14,
			wantMonth: 11,
			wantYear:  2025,
		},
		{
			name:      "1d at 2pm",
			freq:      "1d",
			now:       time.Date(2025, 11, 14, 14, 30, 0, 0, time.UTC),
			wantHour:  0,
			wantDay:   15,
			wantMonth: 11,
			wantYear:  2025,
		},
		{
			name:      "1w on Friday",
			freq:      "1w",
			now:       time.Date(2025, 11, 14, 14, 30, 0, 0, time.UTC), // Friday
			wantHour:  0,
			wantDay:   16, // Sunday
			wantMonth: 11,
			wantYear:  2025,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextExpiryTime(tt.freq, tt.now)
			if err != nil {
				t.Fatalf("NextExpiryTime() error = %v", err)
			}
			if got.Hour() != tt.wantHour || got.Day() != tt.wantDay || got.Month() != tt.wantMonth || got.Year() != tt.wantYear {
				t.Errorf("NextExpiryTime() = %v, want %04d-%02d-%02d %02d:00", got, tt.wantYear, tt.wantMonth, tt.wantDay, tt.wantHour)
			}
		})
	}
}

// Quick manual test for debugging
func TestNextExpiryTimeManual(t *testing.T) {
	now := time.Date(2025, 11, 14, 14, 30, 0, 0, time.UTC)
	fmt.Printf("Now: %s (Friday)\n", now.Format("2006-01-02 15:04 (Monday)"))

	for _, freq := range []string{"1h", "2h", "1d", "2d", "1w"} {
		exp, _ := NextExpiryTime(freq, now)
		fmt.Printf("Freq %s -> Expiry: %s\n", freq, exp.Format("2006-01-02 15:04 (Monday)"))
	}
}
