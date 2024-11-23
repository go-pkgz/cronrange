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
	start      time.Duration // minutes since midnight
	end        time.Duration
	all        bool
	overnight  bool // true if range spans across midnight
	hasSeconds bool // track if the original format included seconds
}

// Field represents a cronrange field that can contain multiple values
type Field struct {
	values map[int]bool
	all    bool
}

// parseRule parses a cronrange rule string and returns a Rule struct or an error if the input is invalid
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

// parseTimeRange parses a time range string in the following formats: HH:MM-HH:MM, HH:MM:SS-HH:MM:SS
// or a single asterisk for all day. Handles ranges that span across midnight.
func parseTimeRange(s string) (TimeRange, error) {
	if s == "*" {
		return TimeRange{all: true}, nil
	}

	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return TimeRange{}, fmt.Errorf("invalid time range format")
	}

	start, hasStartSeconds, err := parseTime(parts[0])
	if err != nil {
		return TimeRange{}, err
	}

	end, hasEndSeconds, err := parseTime(parts[1])
	if err != nil {
		return TimeRange{}, err
	}

	// Check if this is an overnight range
	overnight := false
	if end < start {
		overnight = true
	}

	return TimeRange{
		start:      start,
		end:        end,
		overnight:  overnight,
		hasSeconds: hasStartSeconds || hasEndSeconds,
	}, nil
}

// parseTime parses a time string in the following formats: HH:MM, HH:MM:SS
// Returns the duration, whether seconds were specified, and any error
func parseTime(s string) (time.Duration, bool, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false, fmt.Errorf("invalid time format")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false, err
	}

	seconds := 0
	hasSeconds := len(parts) == 3
	if hasSeconds {
		seconds, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, false, err
		}
	}

	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
		return 0, false, fmt.Errorf("invalid time values")
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, hasSeconds, nil
}

// parseField parses a field string in the following formats: 1,2,3, 1-3,5-6 or a single asterisk for all values.
// The min and max arguments define the range of valid values for the field. The function returns a Field with
// the parsed values or an error if the input is invalid. Values in the Field are stored in a map
// for fast lookup of allowed values.
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

// matches checks if the current time falls within the time range,
// handling ranges that span across midnight
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

	currentTime := time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second

	if r.timeRange.overnight {
		// For overnight ranges (e.g. 23:00-02:00)
		// The time matches if it's:
		// - After or equal to start time (e.g. >= 23:00) OR
		// - Before or equal to end time (e.g. <= 02:00)
		return currentTime >= r.timeRange.start || currentTime <= r.timeRange.end
	}

	// For same-day ranges, time must be between start and end
	return currentTime >= r.timeRange.start && currentTime <= r.timeRange.end
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

	if tr.hasSeconds {
		return fmt.Sprintf("%02d:%02d:%02d-%02d:%02d:%02d",
			startH, startM, startS, endH, endM, endS)
	}
	return fmt.Sprintf("%02d:%02d-%02d:%02d", startH, startM, endH, endM)
}

// String returns the string representation of a Field
func (f Field) String() string {
	if f.all {
		return "*"
	}

	// get all values from the map
	var vals []int
	for v := range f.values {
		vals = append(vals, v)
	}
	if len(vals) == 0 {
		return "*"
	}

	// sort values
	sort.Ints(vals)

	// find ranges and individual values
	var ranges []string
	start := vals[0]
	prev := start

	for i := 1; i < len(vals); i++ {
		if vals[i] != prev+1 {
			// end of a range or single value
			if start == prev {
				ranges = append(ranges, fmt.Sprintf("%d", start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, prev))
			}
			start = vals[i]
		}
		prev = vals[i]
	}

	// handle the last range or value
	if start == prev {
		ranges = append(ranges, fmt.Sprintf("%d", start))
	} else {
		ranges = append(ranges, fmt.Sprintf("%d-%d", start, prev))
	}

	return strings.Join(ranges, ",")
}
