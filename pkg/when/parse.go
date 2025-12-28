package when

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimeExpressionError indicates an invalid time expression.
type TimeExpressionError struct {
	Expression string
	Reason     string
}

func (e *TimeExpressionError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid time expression '%s': %s", e.Expression, e.Reason)
	}
	return fmt.Sprintf("invalid time expression '%s'", e.Expression)
}

var (
	// 12-hour time: 3pm, 3:30pm, 3:30PM, 12am (case-insensitive am/pm)
	time12Regex = regexp.MustCompile(`(?i)^(\d{1,2})(?::(\d{2}))?\s*(am|pm)$`)
	// 24-hour time: 17:00, 05:30, 5:30
	time24Regex = regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
)

// ParseTimeExpression parses a time expression and returns a time.Time on the given reference date.
// Supported formats:
//   - "now" - current time
//   - "3pm", "3:30pm" - 12-hour format
//   - "17:00", "5:30" - 24-hour format
func ParseTimeExpression(expr string, refDate time.Time, loc *time.Location) (time.Time, error) {
	expr = strings.TrimSpace(expr)
	lower := strings.ToLower(expr)

	// Handle "now"
	if lower == "now" {
		return time.Now().In(loc), nil
	}

	var hour, minute int

	// Try 12-hour format
	if matches := time12Regex.FindStringSubmatch(expr); matches != nil {
		h, _ := strconv.Atoi(matches[1])
		m := 0
		if matches[2] != "" {
			m, _ = strconv.Atoi(matches[2])
		}
		ampm := strings.ToLower(matches[3])

		if h < 1 || h > 12 {
			return time.Time{}, &TimeExpressionError{Expression: expr, Reason: "hour must be 1-12 for 12-hour format"}
		}
		if m < 0 || m > 59 {
			return time.Time{}, &TimeExpressionError{Expression: expr, Reason: "minutes must be 0-59"}
		}

		// Convert to 24-hour
		if ampm == "am" {
			if h == 12 {
				hour = 0
			} else {
				hour = h
			}
		} else { // pm
			if h == 12 {
				hour = 12
			} else {
				hour = h + 12
			}
		}
		minute = m
	} else if matches := time24Regex.FindStringSubmatch(expr); matches != nil {
		// Try 24-hour format
		h, _ := strconv.Atoi(matches[1])
		m, _ := strconv.Atoi(matches[2])

		if h < 0 || h > 23 {
			return time.Time{}, &TimeExpressionError{Expression: expr, Reason: "hour must be 0-23 for 24-hour format"}
		}
		if m < 0 || m > 59 {
			return time.Time{}, &TimeExpressionError{Expression: expr, Reason: "minutes must be 0-59"}
		}

		hour = h
		minute = m
	} else {
		return time.Time{}, &TimeExpressionError{Expression: expr}
	}

	// Construct time on reference date in the given location
	year, month, day := refDate.Date()
	return time.Date(year, month, day, hour, minute, 0, 0, loc), nil
}
