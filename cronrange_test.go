package cronrange

import (
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    string // expected String() output
		wantErr bool
	}{
		{
			name: "basic weekday evening",
			expr: "17:20-21:35 1-5 * *",
			want: "17:20-21:35 1-5 * *",
		},
		{
			name: "all weekend",
			expr: "* 0,6 * *",
			want: "* 0,6 * *",
		},
		{
			name: "multiple rules",
			expr: "17:20-21:35 1-5 * *; * 0,6 * *",
			want: "17:20-21:35 1-5 * *; * 0,6 * *",
		},
		{
			name: "specific month days",
			expr: "12:00-13:00 * 1,15 *",
			want: "12:00-13:00 * 1,15 *",
		},
		{
			name: "specific months",
			expr: "09:00-17:00 1-5 * 4-9",
			want: "09:00-17:00 1-5 * 4-9",
		},
		{
			name:    "invalid time format",
			expr:    "1720-2135 1-5 * *",
			wantErr: true,
		},
		{
			name:    "invalid time range",
			expr:    "25:00-26:00 1-5 * *",
			wantErr: true,
		},
		{
			name:    "invalid dow",
			expr:    "17:20-21:35 7 * *",
			wantErr: true,
		},
		{
			name:    "invalid dom",
			expr:    "17:20-21:35 1-5 32 *",
			wantErr: true,
		},
		{
			name:    "invalid month",
			expr:    "17:20-21:35 1-5 * 13",
			wantErr: true,
		},
		{
			name:    "wrong number of fields",
			expr:    "17:20-21:35 1-5 *",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Convert the rules back to string and check
			var gotStr string
			for i, rule := range got {
				if i > 0 {
					gotStr += "; "
				}
				gotStr += rule.String()
			}
			if gotStr != tt.want {
				t.Errorf("Parse() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		time    time.Time
		want    bool
		wantErr bool
	}{
		{
			name: "weekday evening match",
			expr: "17:20-21:35 1-5 * *",
			time: time.Date(2024, 1, 1, 18, 30, 0, 0, time.UTC), // Monday 6:30 PM
			want: true,
		},
		{
			name: "weekday evening non-match time",
			expr: "17:20-21:35 1-5 * *",
			time: time.Date(2024, 1, 1, 16, 30, 0, 0, time.UTC), // Monday 4:30 PM
			want: false,
		},
		{
			name: "weekday evening non-match day",
			expr: "17:20-21:35 1-5 * *",
			time: time.Date(2024, 1, 6, 18, 30, 0, 0, time.UTC), // Saturday 6:30 PM
			want: false,
		},
		{
			name: "weekend all time",
			expr: "* 0,6 * *",
			time: time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC), // Saturday noon
			want: true,
		},
		{
			name: "multiple rules - weekday match",
			expr: "17:20-21:35 1-5 * *; * 0,6 * *",
			time: time.Date(2024, 1, 1, 18, 30, 0, 0, time.UTC), // Monday 6:30 PM
			want: true,
		},
		{
			name: "multiple rules - weekend match",
			expr: "17:20-21:35 1-5 * *; * 0,6 * *",
			time: time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC), // Saturday noon
			want: true,
		},
		{
			name: "multiple rules - weekday non-match",
			expr: "17:20-21:35 1-5 * *; * 0,6 * *",
			time: time.Date(2024, 1, 1, 16, 30, 0, 0, time.UTC), // Monday 4:30 PM
			want: false,
		},
		{
			name: "specific month days match",
			expr: "12:00-13:00 * 1,15 *",
			time: time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC), // 15th at 12:30
			want: true,
		},
		{
			name: "specific month days non-match",
			expr: "12:00-13:00 * 1,15 *",
			time: time.Date(2024, 1, 14, 12, 30, 0, 0, time.UTC), // 14th at 12:30
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := Parse(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			got := Match(rules, tt.time)
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFromReader(t *testing.T) {
	equal := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	tests := []struct {
		name    string
		input   string
		want    []string // expected String() outputs
		wantErr bool
	}{
		{
			name:  "single rule",
			input: "17:20-21:35 1-5 * *",
			want:  []string{"17:20-21:35 1-5 * *"},
		},
		{
			name:  "multiple rules",
			input: "17:20-21:35 1-5 * *\n* 0,6 * *",
			want:  []string{"17:20-21:35 1-5 * *", "* 0,6 * *"},
		},
		{
			name:  "empty input",
			input: "",
			want:  []string{},
		},
		{
			name:    "invalid rule",
			input:   "invalid rule",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rdr := strings.NewReader(tt.input)
			got, err := ParseFromReader(rdr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			var gotStr []string
			for _, rule := range got {
				gotStr = append(gotStr, rule.String())
			}
			if !equal(gotStr, tt.want) {
				t.Errorf("ParseFromReader() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
