package cronrange

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Rule represents a single cronrange rule
type Rule struct {
	timeRange TimeRange
	dow       Field // 0-6 (Sunday = 0)
	dom       Field // 1-31
	month     Field // 1-12
}

// TimeRange represents a time period within a day
type TimeRange struct {
	start time.Duration // minutes since midnight
	end   time.Duration
	all   bool
}

// Field represents a cronrange field that can contain multiple values
type Field struct {
	values map[int]bool
	all    bool
}

func parseRule(rule string) (Rule, error) {
	parts := strings.Fields(rule)
	if len(parts) != 4 {
		return Rule{}, fmt.Errorf("rule must have 4 fields: time dow dom month")
	}

	timeRange, err := parseTimeRange(parts[0])
	if err != nil {
		return Rule{}, err
	}

	dow, err := parseField(parts[1], 0, 6)
	if err != nil {
		return Rule{}, fmt.Errorf("invalid dow: %w", err)
	}

	dom, err := parseField(parts[2], 1, 31)
	if err != nil {
		return Rule{}, fmt.Errorf("invalid dom: %w", err)
	}

	month, err := parseField(parts[3], 1, 12)
	if err != nil {
		return Rule{}, fmt.Errorf("invalid month: %w", err)
	}

	return Rule{
		timeRange: timeRange,
		dow:       dow,
		dom:       dom,
		month:     month,
	}, nil
}

func parseTimeRange(s string) (TimeRange, error) {
	if s == "*" {
		return TimeRange{all: true}, nil
	}

	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return TimeRange{}, fmt.Errorf("invalid time range format")
	}

	start, err := parseTime(parts[0])
	if err != nil {
		return TimeRange{}, err
	}

	end, err := parseTime(parts[1])
	if err != nil {
		return TimeRange{}, err
	}

	return TimeRange{start: start, end: end}, nil
}

// parseTime parses a time string in the following formats: HH:MM, HH-MM-SS
func parseTime(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, fmt.Errorf("invalid time format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	seconds := 0
	if len(parts) == 3 {
		seconds, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, err
		}
	}

	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
		return 0, fmt.Errorf("invalid time values")
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

func parseField(s string, min, max int) (Field, error) {
	if s == "*" {
		return Field{all: true}, nil
	}

	values := make(map[int]bool)
	ranges := strings.Split(s, ",")

	for _, r := range ranges {
		if strings.Contains(r, "-") {
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return Field{}, fmt.Errorf("invalid range format")
			}

			start, err := strconv.Atoi(parts[0])
			if err != nil {
				return Field{}, err
			}

			end, err := strconv.Atoi(parts[1])
			if err != nil {
				return Field{}, err
			}

			if start < min || end > max || start > end {
				return Field{}, fmt.Errorf("values out of range")
			}

			for i := start; i <= end; i++ {
				values[i] = true
			}
			continue
		}

		val, err := strconv.Atoi(r)
		if err != nil {
			return Field{}, err
		}

		if val < min || val > max {
			return Field{}, fmt.Errorf("value out of range")
		}

		values[val] = true
	}

	return Field{values: values}, nil
}

func (r Rule) matches(t time.Time) bool {
	if !r.month.matches(int(t.Month())) {
		return false
	}

	if !r.dom.matches(t.Day()) {
		return false
	}

	if !r.dow.matches(int(t.Weekday())) {
		return false
	}

	if r.timeRange.all {
		return true
	}

	currentMinutes := time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
	return currentMinutes >= r.timeRange.start && currentMinutes <= r.timeRange.end
}

func (f Field) matches(val int) bool {
	return f.all || f.values[val]
}

// String returns the string representation of a Rule
func (r Rule) String() string {
	return fmt.Sprintf("%s %s %s %s",
		r.timeRange.String(),
		r.dow.String(),
		r.dom.String(),
		r.month.String(),
	)
}

// String returns the string representation of a TimeRange
func (tr TimeRange) String() string {
	if tr.all {
		return "*"
	}
	startH := tr.start / time.Hour
	startM := (tr.start % time.Hour) / time.Minute
	startS := (tr.start % time.Minute) / time.Second
	endH := tr.end / time.Hour
	endM := (tr.end % time.Hour) / time.Minute
	endS := (tr.end % time.Minute) / time.Second

	if startS > 0 || endS > 0 {
		return fmt.Sprintf("%02d:%02d:%02d-%02d:%02d:%02d", startH, startM, startS, endH, endM, endS)
	}
	return fmt.Sprintf("%02d:%02d-%02d:%02d", startH, startM, endH, endM)
}

// String returns the string representation of a Field
func (f Field) String() string {
	if f.all {
		return "*"
	}

	// Get all values from the map
	var vals []int
	for v := range f.values {
		vals = append(vals, v)
	}
	if len(vals) == 0 {
		return "*"
	}

	// Sort values
	sort.Ints(vals)

	// Find ranges and individual values
	var ranges []string
	start := vals[0]
	prev := start

	for i := 1; i < len(vals); i++ {
		if vals[i] != prev+1 {
			// End of a range or single value
			if start == prev {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, prev))
			}
			start = vals[i]
		}
		prev = vals[i]
	}

	// Handle the last range or value
	if start == prev {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, prev))
	}

	return strings.Join(ranges, ",")
}
