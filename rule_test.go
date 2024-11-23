package cronrange

import (
	"testing"
	"time"
)

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    string
		wantErr bool
	}{
		{
			name: "all time",
			s:    "*",
			want: "*",
		},
		{
			name: "simple range",
			s:    "09:00-17:00",
			want: "09:00-17:00",
		},
		{
			name: "range with seconds",
			s:    "09:00:11-17:00:22",
			want: "09:00:11-17:00:22",
		},
		{
			name: "evening range",
			s:    "17:20-21:35",
			want: "17:20-21:35",
		},
		{
			name:    "invalid format",
			s:       "9:00to17:00",
			wantErr: true,
		},
		{
			name:    "invalid hour",
			s:       "24:00-17:00",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			s:       "09:60-17:00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeRange(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.String() != tt.want {
				t.Errorf("parseTimeRange() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestParseField(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		min     int
		max     int
		want    string
		wantErr bool
	}{
		{
			name: "all values",
			s:    "*",
			min:  0,
			max:  6,
			want: "*",
		},
		{
			name: "single value",
			s:    "5",
			min:  0,
			max:  6,
			want: "5",
		},
		{
			name: "value list",
			s:    "1,3,5",
			min:  0,
			max:  6,
			want: "1,3,5",
		},
		{
			name: "range",
			s:    "1-5",
			min:  0,
			max:  6,
			want: "1-5",
		},
		{
			name: "complex range",
			s:    "1-3,5-6",
			min:  0,
			max:  6,
			want: "1-3,5-6",
		},
		{
			name:    "out of range",
			s:       "7",
			min:     0,
			max:     6,
			wantErr: true,
		},
		{
			name:    "invalid range",
			s:       "5-3",
			min:     0,
			max:     6,
			wantErr: true,
		},
		{
			name:    "invalid format",
			s:       "a-b",
			min:     0,
			max:     6,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseField(tt.s, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.String() != tt.want {
				t.Errorf("parseField() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestTimeRangeString(t *testing.T) {
	tests := []struct {
		name string
		tr   TimeRange
		want string
	}{
		{
			name: "all time",
			tr:   TimeRange{all: true},
			want: "*",
		},
		{
			name: "simple range",
			tr: TimeRange{
				start: 9*time.Hour + 0*time.Minute,
				end:   17*time.Hour + 0*time.Minute,
			},
			want: "09:00-17:00",
		},
		{
			name: "complex range",
			tr: TimeRange{
				start: 17*time.Hour + 20*time.Minute,
				end:   21*time.Hour + 35*time.Minute,
			},
			want: "17:20-21:35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("TimeRange.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOvernightTimeRange(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"overnight range", "23:00-02:00", "23:00-02:00"},
		{"overnight range with seconds", "22:30:00-04:15:00", "22:30:00-04:15:00"},
		{"regular range", "09:00-17:00", "09:00-17:00"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			timeRange, err := parseTimeRange(c.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got := timeRange.String(); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}

			if c.input == "23:00-02:00" {
				if !timeRange.overnight {
					t.Error("overnight should be true for 23:00-02:00")
				}

				rule := Rule{
					timeRange: timeRange,
					dow:       Field{all: true},
					dom:       Field{all: true},
					month:     Field{all: true},
				}

				baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				times := []struct {
					t    time.Time
					name string
					want bool
				}{
					{baseTime.Add(-time.Hour), "23:00", true},
					{baseTime, "00:00", true},
					{baseTime.Add(time.Hour), "01:00", true},
					{baseTime.Add(2 * time.Hour), "02:00", true},
					{baseTime.Add(3 * time.Hour), "03:00", false},
					{baseTime.Add(-2 * time.Hour), "22:00", false},
				}

				for _, tc := range times {
					if got := rule.matches(tc.t); got != tc.want {
						t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
					}
				}
			}
		})
	}
}

func TestMatches(t *testing.T) {
	cases := []struct {
		name      string
		rule      string
		times     []time.Time
		wantMatch []bool
	}{
		{
			name: "simple range",
			rule: "09:00-17:00 * * *",
			times: []time.Time{
				time.Date(2024, 1, 1, 8, 59, 59, 0, time.UTC), // before range
				time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),   // start of range
				time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),  // middle
				time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),  // end of range
				time.Date(2024, 1, 1, 17, 0, 1, 0, time.UTC),  // after range
			},
			wantMatch: []bool{false, true, true, true, false},
		},
		{
			name: "overnight range",
			rule: "23:00-02:00 * * *",
			times: []time.Time{
				time.Date(2024, 1, 1, 22, 59, 59, 0, time.UTC), // just before
				time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC),   // start
				time.Date(2024, 1, 1, 23, 59, 59, 0, time.UTC), // before midnight
				time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),    // midnight
				time.Date(2024, 1, 2, 1, 30, 0, 0, time.UTC),   // middle
				time.Date(2024, 1, 2, 2, 0, 0, 0, time.UTC),    // end
				time.Date(2024, 1, 2, 2, 0, 1, 0, time.UTC),    // just after
			},
			wantMatch: []bool{false, true, true, true, true, true, false},
		},
		{
			name: "specific days",
			rule: "10:00-12:00 1,3,5 * *", // Mon,Wed,Fri
			times: []time.Time{
				time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), // Monday
				time.Date(2024, 1, 2, 11, 0, 0, 0, time.UTC), // Tuesday
				time.Date(2024, 1, 3, 11, 0, 0, 0, time.UTC), // Wednesday
			},
			wantMatch: []bool{true, false, true},
		},
		{
			name: "specific months",
			rule: "* * * 3,6,9,12", // Mar,Jun,Sep,Dec
			times: []time.Time{
				time.Date(2024, 2, 1, 11, 0, 0, 0, time.UTC), // February
				time.Date(2024, 3, 1, 11, 0, 0, 0, time.UTC), // March
				time.Date(2024, 4, 1, 11, 0, 0, 0, time.UTC), // April
			},
			wantMatch: []bool{false, true, false},
		},
		{
			name: "complex range",
			rule: "09:00-17:00 1-5 1-7 3,6,9,12", // weekdays, first week, specific months
			times: []time.Time{
				time.Date(2024, 3, 1, 13, 0, 0, 0, time.UTC), // Fri, Mar 1
				time.Date(2024, 3, 2, 13, 0, 0, 0, time.UTC), // Sat, Mar 2
				time.Date(2024, 3, 4, 13, 0, 0, 0, time.UTC), // Mon, Mar 4
				time.Date(2024, 3, 8, 13, 0, 0, 0, time.UTC), // Fri, Mar 8
				time.Date(2024, 4, 1, 13, 0, 0, 0, time.UTC), // Mon, Apr 1
			},
			wantMatch: []bool{true, false, true, false, false},
		},
		{
			name: "all wildcards",
			rule: "* * * *",
			times: []time.Time{
				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC),
				time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
			wantMatch: []bool{true, true, true},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rule, err := parseRule(c.rule)
			if err != nil {
				t.Fatalf("failed to parse rule %q: %v", c.rule, err)
			}

			if len(c.times) != len(c.wantMatch) {
				t.Fatal("times and wantMatch slices must have equal length")
			}

			for i, tm := range c.times {
				got := rule.matches(tm)
				if got != c.wantMatch[i] {
					t.Errorf("time %v: got %v, want %v", tm.Format("2006-01-02 15:04:05"), got, c.wantMatch[i])
				}
			}
		})
	}
}
