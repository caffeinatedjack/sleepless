package when

import (
	"testing"
	"time"
)

func TestParseTimeExpression(t *testing.T) {
	// Use a fixed reference date
	refDate := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	loc := time.UTC

	tests := []struct {
		name     string
		expr     string
		wantHour int
		wantMin  int
		wantErr  bool
	}{
		// 12-hour format
		{"12h basic pm", "3pm", 15, 0, false},
		{"12h with minutes pm", "3:30pm", 15, 30, false},
		{"12h basic am", "9am", 9, 0, false},
		{"12h with minutes am", "9:15am", 9, 15, false},
		{"12h noon", "12pm", 12, 0, false},
		{"12h midnight", "12am", 0, 0, false},
		{"12h uppercase", "3PM", 15, 0, false},
		{"12h mixed case", "3Pm", 15, 0, false},

		// 24-hour format
		{"24h basic", "17:00", 17, 0, false},
		{"24h with minutes", "05:30", 5, 30, false},
		{"24h midnight", "00:00", 0, 0, false},
		{"24h noon", "12:00", 12, 0, false},
		{"24h single digit hour", "9:30", 9, 30, false},

		// Invalid formats
		{"invalid hour 12h", "13pm", 0, 0, true},
		{"invalid hour 24h", "25:00", 0, 0, true},
		{"invalid minutes", "3:60pm", 0, 0, true},
		{"invalid format", "three o'clock", 0, 0, true},
		{"empty", "", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimeExpression(tt.expr, refDate, loc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Hour() != tt.wantHour {
					t.Errorf("ParseTimeExpression() hour = %d, want %d", got.Hour(), tt.wantHour)
				}
				if got.Minute() != tt.wantMin {
					t.Errorf("ParseTimeExpression() minute = %d, want %d", got.Minute(), tt.wantMin)
				}
			}
		})
	}
}

func TestParseTimeExpression_Now(t *testing.T) {
	refDate := time.Now()
	loc := time.UTC

	got, err := ParseTimeExpression("now", refDate, loc)
	if err != nil {
		t.Errorf("ParseTimeExpression(now) error = %v", err)
		return
	}

	// Should be within a second of now
	diff := time.Since(got)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("ParseTimeExpression(now) returned time too far from now: %v", diff)
	}
}

func TestParseTimeExpression_PreservesDate(t *testing.T) {
	refDate := time.Date(2025, 12, 25, 10, 0, 0, 0, time.UTC)
	loc := time.UTC

	got, err := ParseTimeExpression("3pm", refDate, loc)
	if err != nil {
		t.Errorf("ParseTimeExpression() error = %v", err)
		return
	}

	if got.Year() != 2025 || got.Month() != 12 || got.Day() != 25 {
		t.Errorf("ParseTimeExpression() date = %s, want 2025-12-25", got.Format("2006-01-02"))
	}
}
