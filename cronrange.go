// Package cronrange provides a crontab-like format for expressing time ranges.
// Unlike traditional crontab that defines specific moments in time, cronrange
// defines time periods when something should be active.
//
// Format:
//
//	time dow dom month
//
// Where:
//   - time:  Time range in 24h format (HH:MM-HH:MM) or * for all day
//   - dow:   Day of week (0-6, where 0=Sunday)
//   - dom:   Day of month (1-31)
//   - month: Month (1-12)
//
// Each field (except time) supports single values, lists (1,3,5), ranges (1-5)
// and asterisk (*) for any/all values. Multiple rules can be combined using semicolons.
//
// Examples:
//
//	17:20-21:35 1-5 * *          # Weekdays from 5:20 PM to 9:35 PM
//	* 0,6 * *                    # All day on weekends
//	09:00-17:00 1-5 * 4-9        # Weekdays 9 AM to 5 PM, April through September
//	12:00-13:00 * 1,15 *         # Noon-1 PM on 1st and 15th of every month
package cronrange

import (
	"fmt"
	"strings"
	"time"
)

// Parse parses a cronrange expression and returns a Rule slice
func Parse(expr string) ([]Rule, error) {
	rules := strings.Split(expr, ";")
	result := make([]Rule, 0, len(rules))

	for _, r := range rules {
		rule, err := parseRule(strings.TrimSpace(r))
		if err != nil {
			return nil, fmt.Errorf("invalid rule '%s': %w", r, err)
		}
		result = append(result, rule)
	}

	return result, nil
}

// Match checks if the given time matches any of the rules
func Match(rules []Rule, t time.Time) bool {
	for _, rule := range rules {
		if rule.matches(t) {
			return true
		}
	}
	return false
}
