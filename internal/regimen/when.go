package regimen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
	"gitlab.com/caffeinatedjack/sleepless/pkg/when"
)

var (
	whenJSON  bool
	whenShort bool
	when24h   bool
	when12h   bool
	whenDate  bool
)

// JSON output types
type WhenZoneOutput struct {
	Label string `json:"label"`
	Zone  string `json:"zone"`
	Time  string `json:"time"`
}

type WhenOutput struct {
	Reference WhenZoneOutput   `json:"reference"`
	Zones     []WhenZoneOutput `json:"zones"`
}

type WhenConversionOutput struct {
	From WhenZoneOutput `json:"from"`
	To   WhenZoneOutput `json:"to"`
}

var whenCmd = &cobra.Command{
	Use:   "when [time-expr] [in zone] [to zone]",
	Short: "World clock and time conversion",
	Long: `World clock and time conversion utilities.

Commands:
    when                       Show current time across configured zones
    when <time>                Convert local time to configured zones
    when <time> in <zone>      Convert time from zone to configured zones
    when <time> in <z1> to <z2> Direct conversion between two zones

    when add <alias> <zone>    Add a timezone alias
    when remove <alias>        Remove a timezone alias
    when zones [query]         List available timezones

    when diff <zone> <zone>    Show time difference between zones
    when until <time>          Show duration until a time
    when overlap <zone> <zone> Find overlapping work hours

Time formats:
    now, 3pm, 3:30pm, 17:00, 05:30

Zone formats:
    - IANA zones: America/New_York, Europe/London
    - Abbreviations: UTC, PST, JST
    - Cities: London, Tokyo, NYC
    - Your aliases: work, home

Examples:
    regimen when
    regimen when 3pm
    regimen when 3pm in London
    regimen when 3pm in London to Tokyo
    regimen when add work America/New_York
    regimen when zones europe`,
	Run: runWhen,
}

func init() {
	rootCmd.AddCommand(whenCmd)

	whenCmd.PersistentFlags().BoolVar(&whenJSON, "json", false, "Output as JSON")
	whenCmd.PersistentFlags().BoolVar(&whenShort, "short", false, "Compact single-line output")
	whenCmd.PersistentFlags().BoolVar(&when24h, "24h", false, "Use 24-hour time format (default)")
	whenCmd.PersistentFlags().BoolVar(&when12h, "12h", false, "Use 12-hour time format")
	whenCmd.PersistentFlags().BoolVar(&whenDate, "date", false, "Include date in output")

	whenCmd.AddCommand(whenAddCmd)
	whenCmd.AddCommand(whenRemoveCmd)
	whenCmd.AddCommand(whenZonesCmd)
	whenCmd.AddCommand(whenDiffCmd)
	whenCmd.AddCommand(whenUntilCmd)
	whenCmd.AddCommand(whenOverlapCmd)
}

// formatTime formats a time according to the current flags.
func formatTime(t time.Time, refDate time.Time) string {
	var format string

	if when12h {
		format = "3:04 PM"
	} else {
		format = "15:04"
	}

	if whenDate {
		format = "2006-01-02 " + format
	}

	result := t.Format(format)

	// If not showing date but days differ from reference, indicate it
	if !whenDate {
		refYear, refMonth, refDay := refDate.Date()
		tYear, tMonth, tDay := t.Date()
		if tYear != refYear || tMonth != refMonth || tDay != refDay {
			diff := int(t.Sub(time.Date(refYear, refMonth, refDay, 0, 0, 0, 0, t.Location())).Hours() / 24)
			if diff > 0 {
				result += fmt.Sprintf(" (+%dd)", diff)
			} else if diff < 0 {
				result += fmt.Sprintf(" (%dd)", diff)
			}
		}
	}

	return result
}

func runWhen(cmd *cobra.Command, args []string) {
	cfg, err := when.LoadConfig()
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Parse arguments to detect "in" and "to" clauses
	var timeExpr string
	var inZone, toZone string
	parseArgs(args, &timeExpr, &inZone, &toZone)

	// Determine reference location
	var refLoc *time.Location
	var refLabel string
	if inZone != "" {
		resolved, err := when.ResolveZone(inZone, cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		refLoc = resolved.Location
		refLabel = resolved.Label
	} else {
		refLoc = time.Local
		refLabel = "Local"
	}

	// Parse time expression
	now := time.Now()
	var refTime time.Time
	if timeExpr == "" || strings.ToLower(timeExpr) == "now" {
		refTime = now.In(refLoc)
	} else {
		parsed, err := when.ParseTimeExpression(timeExpr, now, refLoc)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		refTime = parsed
	}

	// Handle direct conversion (to <zone>)
	if toZone != "" {
		resolved, err := when.ResolveZone(toZone, cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		outputDirectConversion(refTime, refLabel, resolved)
		return
	}

	// Show time across configured zones
	outputWorldClock(refTime, refLabel, cfg)
}

func parseArgs(args []string, timeExpr, inZone, toZone *string) {
	// Look for "in" and "to" keywords
	inIdx := -1
	toIdx := -1
	for i, arg := range args {
		lower := strings.ToLower(arg)
		if lower == "in" && inIdx == -1 {
			inIdx = i
		} else if lower == "to" && toIdx == -1 {
			toIdx = i
		}
	}

	// Extract time expression (everything before "in" or all args if no "in")
	if inIdx > 0 {
		*timeExpr = strings.Join(args[:inIdx], " ")
	} else if inIdx == -1 && toIdx == -1 && len(args) > 0 {
		*timeExpr = strings.Join(args, " ")
	} else if inIdx == 0 {
		// "in" is first arg, no time expression
		*timeExpr = "now"
	}

	// Extract "in" zone
	if inIdx >= 0 {
		endIdx := len(args)
		if toIdx > inIdx {
			endIdx = toIdx
		}
		if inIdx+1 < endIdx {
			*inZone = strings.Join(args[inIdx+1:endIdx], " ")
		}
	}

	// Extract "to" zone
	if toIdx >= 0 && toIdx+1 < len(args) {
		*toZone = strings.Join(args[toIdx+1:], " ")
	}
}

func outputDirectConversion(refTime time.Time, refLabel string, to *when.ResolvedZone) {
	toTime := refTime.In(to.Location)

	if whenJSON {
		output := WhenConversionOutput{
			From: WhenZoneOutput{
				Label: refLabel,
				Zone:  refTime.Location().String(),
				Time:  refTime.Format(time.RFC3339),
			},
			To: WhenZoneOutput{
				Label: to.Label,
				Zone:  to.Zone,
				Time:  toTime.Format(time.RFC3339),
			},
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return
	}

	if whenShort {
		fmt.Printf("%s %s -> %s %s\n",
			formatTime(refTime, refTime), refLabel,
			formatTime(toTime, refTime), to.Label)
		return
	}

	// Standard output
	fromStyle := lipgloss.NewStyle().Foreground(ui.Blue)
	toStyle := lipgloss.NewStyle().Foreground(ui.Green)

	fmt.Println()
	fmt.Printf("  %s  %s\n", fromStyle.Render(formatTime(refTime, refTime)), refLabel)
	fmt.Printf("  %s  %s\n", toStyle.Render(formatTime(toTime, refTime)), to.Label)
	fmt.Println()
}

func outputWorldClock(refTime time.Time, refLabel string, cfg *when.Config) {
	// Build list of zones to display
	var zones []*when.ResolvedZone

	// Always include UTC if nothing configured
	if len(cfg.Configured) == 0 {
		if resolved, err := when.ResolveZone("UTC", cfg); err == nil {
			zones = append(zones, resolved)
		}
	} else {
		for _, z := range cfg.Configured {
			resolved, err := when.ResolveZone(z, cfg)
			if err != nil {
				// Skip invalid zones with a warning
				continue
			}
			zones = append(zones, resolved)
		}
	}

	if whenJSON {
		output := WhenOutput{
			Reference: WhenZoneOutput{
				Label: refLabel,
				Zone:  refTime.Location().String(),
				Time:  refTime.Format(time.RFC3339),
			},
			Zones: make([]WhenZoneOutput, len(zones)),
		}
		for i, z := range zones {
			t := refTime.In(z.Location)
			output.Zones[i] = WhenZoneOutput{
				Label: z.Label,
				Zone:  z.Zone,
				Time:  t.Format(time.RFC3339),
			}
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return
	}

	if whenShort {
		parts := []string{fmt.Sprintf("%s %s", formatTime(refTime, refTime), refLabel)}
		for _, z := range zones {
			t := refTime.In(z.Location)
			parts = append(parts, fmt.Sprintf("%s %s", formatTime(t, refTime), z.Label))
		}
		fmt.Println(strings.Join(parts, " | "))
		return
	}

	// Standard output
	refStyle := lipgloss.NewStyle().Foreground(ui.Blue).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(ui.DimColor)

	fmt.Println()
	fmt.Printf("  %s  %s\n", refStyle.Render(formatTime(refTime, refTime)), refLabel)
	fmt.Println()

	for _, z := range zones {
		t := refTime.In(z.Location)
		fmt.Printf("  %s  %s\n", formatTime(t, refTime), labelStyle.Render(z.Label))
	}
	fmt.Println()
}

// whenAddCmd adds a timezone alias
var whenAddCmd = &cobra.Command{
	Use:   "add <alias> <zone>",
	Short: "Add a timezone alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		zone := args[1]

		cfg, err := when.LoadConfig()
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			os.Exit(1)
		}

		// Validate the zone
		resolved, err := when.ResolveZone(zone, cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		cfg.AddAlias(alias, resolved.Zone)

		if err := cfg.Save(); err != nil {
			ui.Error(fmt.Sprintf("Failed to save config: %v", err))
			os.Exit(1)
		}

		ui.Success(fmt.Sprintf("Added alias '%s' -> %s", alias, resolved.Zone))
	},
}

// whenRemoveCmd removes a timezone alias
var whenRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a timezone alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		cfg, err := when.LoadConfig()
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			os.Exit(1)
		}

		if !cfg.RemoveAlias(alias) {
			ui.Error(fmt.Sprintf("Alias '%s' not found", alias))
			os.Exit(1)
		}

		if err := cfg.Save(); err != nil {
			ui.Error(fmt.Sprintf("Failed to save config: %v", err))
			os.Exit(1)
		}

		ui.Success(fmt.Sprintf("Removed alias '%s'", alias))
	},
}

// whenZonesCmd lists available timezones
var whenZonesCmd = &cobra.Command{
	Use:   "zones [query]",
	Short: "List available timezones",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		zones, err := when.ListZones()
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to enumerate zones: %v", err))
			os.Exit(1)
		}

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		filtered := when.FilterZones(zones, query)

		if whenJSON {
			data, _ := json.MarshalIndent(filtered, "", "  ")
			fmt.Println(string(data))
			return
		}

		for _, z := range filtered {
			fmt.Println(z)
		}
	},
}

// whenDiffCmd shows the difference between two timezones
var whenDiffCmd = &cobra.Command{
	Use:   "diff <zone-a> <zone-b>",
	Short: "Show time difference between two zones",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := when.LoadConfig()
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			os.Exit(1)
		}

		zoneA, err := when.ResolveZone(args[0], cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		zoneB, err := when.ResolveZone(args[1], cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		now := time.Now()
		timeA := now.In(zoneA.Location)
		timeB := now.In(zoneB.Location)

		_, offsetA := timeA.Zone()
		_, offsetB := timeB.Zone()

		diffSeconds := offsetA - offsetB
		diffHours := float64(diffSeconds) / 3600

		if whenJSON {
			output := map[string]interface{}{
				"zone_a":           zoneA.Zone,
				"zone_b":           zoneB.Zone,
				"difference_hours": diffHours,
			}
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
			return
		}

		var relation string
		absDiff := diffHours
		if absDiff < 0 {
			absDiff = -absDiff
		}

		if diffHours > 0 {
			relation = "ahead of"
		} else if diffHours < 0 {
			relation = "behind"
		} else {
			fmt.Printf("%s and %s are in the same timezone offset\n", zoneA.Label, zoneB.Label)
			return
		}

		hourStr := "hours"
		if absDiff == 1 {
			hourStr = "hour"
		}

		// Handle half-hour offsets
		if absDiff == float64(int(absDiff)) {
			fmt.Printf("%s is %.0f %s %s %s\n", zoneA.Label, absDiff, hourStr, relation, zoneB.Label)
		} else {
			fmt.Printf("%s is %.1f %s %s %s\n", zoneA.Label, absDiff, hourStr, relation, zoneB.Label)
		}
	},
}

// whenUntilCmd shows duration until a time
var whenUntilCmd = &cobra.Command{
	Use:   "until <time-expr>",
	Short: "Show duration until a time",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		timeExpr := strings.Join(args, " ")

		now := time.Now()
		target, err := when.ParseTimeExpression(timeExpr, now, time.Local)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		// If target is in the past, use tomorrow
		if target.Before(now) || target.Equal(now) {
			target = target.AddDate(0, 0, 1)
		}

		duration := target.Sub(now)

		if whenJSON {
			output := map[string]interface{}{
				"target":           target.Format(time.RFC3339),
				"duration_seconds": duration.Seconds(),
				"duration_human":   formatDuration(duration),
			}
			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("%s until %s\n", formatDuration(duration), formatTime(target, now))
	},
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// whenOverlapCmd finds overlapping work hours
var whenOverlapCmd = &cobra.Command{
	Use:   "overlap <zone-a> <zone-b>",
	Short: "Find overlapping work hours between two zones",
	Long: `Find overlapping work hours between two zones.

Uses default work hours of 09:00-17:00 in each zone.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := when.LoadConfig()
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to load config: %v", err))
			os.Exit(1)
		}

		zoneA, err := when.ResolveZone(args[0], cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		zoneB, err := when.ResolveZone(args[1], cfg)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		// Work hours: 09:00-17:00 in each zone
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

		// Zone A work hours in UTC
		workStartA := time.Date(today.Year(), today.Month(), today.Day(), 9, 0, 0, 0, zoneA.Location)
		workEndA := time.Date(today.Year(), today.Month(), today.Day(), 17, 0, 0, 0, zoneA.Location)

		// Zone B work hours in UTC
		workStartB := time.Date(today.Year(), today.Month(), today.Day(), 9, 0, 0, 0, zoneB.Location)
		workEndB := time.Date(today.Year(), today.Month(), today.Day(), 17, 0, 0, 0, zoneB.Location)

		// Find overlap
		overlapStart := workStartA
		if workStartB.After(workStartA) {
			overlapStart = workStartB
		}

		overlapEnd := workEndA
		if workEndB.Before(workEndA) {
			overlapEnd = workEndB
		}

		if whenJSON {
			output := map[string]interface{}{
				"zone_a": map[string]string{
					"zone":       zoneA.Zone,
					"work_start": workStartA.Format(time.RFC3339),
					"work_end":   workEndA.Format(time.RFC3339),
				},
				"zone_b": map[string]string{
					"zone":       zoneB.Zone,
					"work_start": workStartB.Format(time.RFC3339),
					"work_end":   workEndB.Format(time.RFC3339),
				},
			}

			if overlapEnd.After(overlapStart) {
				output["overlap"] = map[string]interface{}{
					"start":          overlapStart.Format(time.RFC3339),
					"end":            overlapEnd.Format(time.RFC3339),
					"duration_hours": overlapEnd.Sub(overlapStart).Hours(),
				}
			} else {
				output["overlap"] = nil
			}

			data, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(data))
			return
		}

		if !overlapEnd.After(overlapStart) {
			ui.Warning(fmt.Sprintf("No overlapping work hours between %s and %s", zoneA.Label, zoneB.Label))
			return
		}

		duration := overlapEnd.Sub(overlapStart)
		hours := int(duration.Hours())

		fmt.Println()
		fmt.Printf("  %s work hours: 09:00-17:00\n", zoneA.Label)
		fmt.Printf("  %s work hours: 09:00-17:00\n", zoneB.Label)
		fmt.Println()

		overlapStyle := lipgloss.NewStyle().Foreground(ui.Green).Bold(true)
		fmt.Printf("  Overlap: %s (%d hours)\n",
			overlapStyle.Render(fmt.Sprintf("%s - %s %s",
				overlapStart.In(zoneA.Location).Format("15:04"),
				overlapEnd.In(zoneA.Location).Format("15:04"),
				zoneA.Label)),
			hours)
		fmt.Printf("           %s\n",
			overlapStyle.Render(fmt.Sprintf("%s - %s %s",
				overlapStart.In(zoneB.Location).Format("15:04"),
				overlapEnd.In(zoneB.Location).Format("15:04"),
				zoneB.Label)))
		fmt.Println()
	},
}
