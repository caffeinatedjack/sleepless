package when

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Common timezone abbreviations mapped to IANA zones.
// Only includes unambiguous abbreviations.
var abbreviations = map[string]string{
	"UTC":  "UTC",
	"GMT":  "GMT",
	"EST":  "America/New_York",
	"EDT":  "America/New_York",
	"CST":  "America/Chicago", // US Central; China Standard Time users should use "Asia/Shanghai"
	"CDT":  "America/Chicago",
	"MST":  "America/Denver",
	"MDT":  "America/Denver",
	"PST":  "America/Los_Angeles",
	"PDT":  "America/Los_Angeles",
	"HST":  "Pacific/Honolulu",
	"AKST": "America/Anchorage",
	"AKDT": "America/Anchorage",
	"JST":  "Asia/Tokyo",
	"KST":  "Asia/Seoul",
	"IST":  "Asia/Kolkata",
	"HKT":  "Asia/Hong_Kong",
	"SGT":  "Asia/Singapore",
	"AEST": "Australia/Sydney",
	"AEDT": "Australia/Sydney",
	"AWST": "Australia/Perth",
	"NZST": "Pacific/Auckland",
	"NZDT": "Pacific/Auckland",
	"CET":  "Europe/Paris",
	"CEST": "Europe/Paris",
	"WET":  "Europe/London",
	"WEST": "Europe/London",
	"EET":  "Europe/Helsinki",
	"EEST": "Europe/Helsinki",
}

// Common city names mapped to IANA zones.
var cities = map[string]string{
	// Americas
	"new york":      "America/New_York",
	"nyc":           "America/New_York",
	"los angeles":   "America/Los_Angeles",
	"la":            "America/Los_Angeles",
	"chicago":       "America/Chicago",
	"denver":        "America/Denver",
	"phoenix":       "America/Phoenix",
	"seattle":       "America/Los_Angeles",
	"san francisco": "America/Los_Angeles",
	"sf":            "America/Los_Angeles",
	"miami":         "America/New_York",
	"boston":        "America/New_York",
	"toronto":       "America/Toronto",
	"vancouver":     "America/Vancouver",
	"mexico city":   "America/Mexico_City",
	"sao paulo":     "America/Sao_Paulo",

	// Europe
	"london":     "Europe/London",
	"paris":      "Europe/Paris",
	"berlin":     "Europe/Berlin",
	"amsterdam":  "Europe/Amsterdam",
	"brussels":   "Europe/Brussels",
	"madrid":     "Europe/Madrid",
	"rome":       "Europe/Rome",
	"vienna":     "Europe/Vienna",
	"zurich":     "Europe/Zurich",
	"stockholm":  "Europe/Stockholm",
	"oslo":       "Europe/Oslo",
	"copenhagen": "Europe/Copenhagen",
	"helsinki":   "Europe/Helsinki",
	"dublin":     "Europe/Dublin",
	"lisbon":     "Europe/Lisbon",
	"moscow":     "Europe/Moscow",
	"warsaw":     "Europe/Warsaw",
	"prague":     "Europe/Prague",
	"budapest":   "Europe/Budapest",
	"athens":     "Europe/Athens",
	"istanbul":   "Europe/Istanbul",

	// Asia
	"tokyo":        "Asia/Tokyo",
	"osaka":        "Asia/Tokyo",
	"seoul":        "Asia/Seoul",
	"beijing":      "Asia/Shanghai",
	"shanghai":     "Asia/Shanghai",
	"hong kong":    "Asia/Hong_Kong",
	"singapore":    "Asia/Singapore",
	"taipei":       "Asia/Taipei",
	"mumbai":       "Asia/Kolkata",
	"delhi":        "Asia/Kolkata",
	"bangalore":    "Asia/Kolkata",
	"bangkok":      "Asia/Bangkok",
	"jakarta":      "Asia/Jakarta",
	"manila":       "Asia/Manila",
	"kuala lumpur": "Asia/Kuala_Lumpur",
	"dubai":        "Asia/Dubai",
	"tel aviv":     "Asia/Jerusalem",
	"jerusalem":    "Asia/Jerusalem",

	// Oceania
	"sydney":    "Australia/Sydney",
	"melbourne": "Australia/Melbourne",
	"brisbane":  "Australia/Brisbane",
	"perth":     "Australia/Perth",
	"auckland":  "Pacific/Auckland",

	// Africa
	"cairo":        "Africa/Cairo",
	"johannesburg": "Africa/Johannesburg",
	"lagos":        "Africa/Lagos",
	"nairobi":      "Africa/Nairobi",
}

// ResolvedZone represents a resolved timezone.
type ResolvedZone struct {
	// Label is the display label (alias name if used, otherwise the zone string).
	Label string
	// Zone is the IANA zone string.
	Zone string
	// Location is the loaded time.Location.
	Location *time.Location
}

// AmbiguousZoneError indicates a zone token matched multiple possible zones.
type AmbiguousZoneError struct {
	Token   string
	Options []string
}

func (e *AmbiguousZoneError) Error() string {
	return fmt.Sprintf("ambiguous zone '%s': could be %s", e.Token, strings.Join(e.Options, ", "))
}

// UnknownZoneError indicates a zone token could not be resolved.
type UnknownZoneError struct {
	Token string
}

func (e *UnknownZoneError) Error() string {
	return fmt.Sprintf("unknown zone: %s", e.Token)
}

// ResolveZone resolves a zone token to a ResolvedZone.
// Resolution order: alias -> IANA -> abbreviation -> city
func ResolveZone(token string, cfg *Config) (*ResolvedZone, error) {
	// 1. Check aliases first
	if cfg != nil {
		if zone, ok := cfg.Aliases[token]; ok {
			loc, err := time.LoadLocation(zone)
			if err != nil {
				return nil, fmt.Errorf("invalid zone in alias '%s': %w", token, err)
			}
			return &ResolvedZone{Label: token, Zone: zone, Location: loc}, nil
		}
	}

	// 2. Try as IANA zone
	if loc, err := time.LoadLocation(token); err == nil {
		return &ResolvedZone{Label: token, Zone: token, Location: loc}, nil
	}

	// 3. Check abbreviations (case-insensitive)
	upper := strings.ToUpper(token)
	if zone, ok := abbreviations[upper]; ok {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			return nil, fmt.Errorf("failed to load zone for abbreviation '%s': %w", token, err)
		}
		return &ResolvedZone{Label: upper, Zone: zone, Location: loc}, nil
	}

	// 4. Check city names (case-insensitive)
	lower := strings.ToLower(token)
	if zone, ok := cities[lower]; ok {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			return nil, fmt.Errorf("failed to load zone for city '%s': %w", token, err)
		}
		return &ResolvedZone{Label: token, Zone: zone, Location: loc}, nil
	}

	return nil, &UnknownZoneError{Token: token}
}

// ListZones enumerates available IANA timezone names from the system.
func ListZones() ([]string, error) {
	zoneinfoDir := os.Getenv("ZONEINFO")
	if zoneinfoDir == "" {
		// Try common locations
		candidates := []string{
			"/usr/share/zoneinfo",
			"/usr/lib/zoneinfo",
			"/usr/share/lib/zoneinfo",
			"/etc/zoneinfo",
		}
		for _, dir := range candidates {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				zoneinfoDir = dir
				break
			}
		}
	}

	if zoneinfoDir == "" {
		return nil, fmt.Errorf("could not locate zoneinfo directory")
	}

	var zones []string
	err := filepath.Walk(zoneinfoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Skip directories and special files
		if info.IsDir() {
			// Skip certain directories that don't contain timezone data
			name := info.Name()
			if name == "posix" || name == "right" {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from zoneinfo root
		relPath, err := filepath.Rel(zoneinfoDir, path)
		if err != nil {
			return nil
		}

		// Skip files without a slash (they're usually special files like "leap-seconds.list")
		// and files that start with a dot
		if !strings.Contains(relPath, string(filepath.Separator)) && !strings.Contains(relPath, "/") {
			// Single-level files like "UTC", "GMT" are valid
			if len(relPath) <= 5 {
				zones = append(zones, relPath)
			}
			return nil
		}

		// Skip files starting with dot or containing uppercase only in first segment
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 0 && strings.HasPrefix(parts[0], ".") {
			return nil
		}

		// Validate it's a real timezone by trying to load it
		if _, err := time.LoadLocation(relPath); err == nil {
			zones = append(zones, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to enumerate zones: %w", err)
	}

	sort.Strings(zones)
	return zones, nil
}

// FilterZones returns zones that contain the query string (case-insensitive).
func FilterZones(zones []string, query string) []string {
	if query == "" {
		return zones
	}

	query = strings.ToLower(query)
	var filtered []string
	for _, z := range zones {
		if strings.Contains(strings.ToLower(z), query) {
			filtered = append(filtered, z)
		}
	}
	return filtered
}
