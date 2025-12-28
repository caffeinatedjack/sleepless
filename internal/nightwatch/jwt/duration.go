package jwt

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses duration strings with support for days (e.g., "7d", "1h30m").
func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	// Handle days suffix (not in stdlib)
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	if dur < 0 {
		return 0, fmt.Errorf("duration must be positive: %s", s)
	}
	return dur, nil
}
