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
