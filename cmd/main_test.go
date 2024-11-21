package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	// build the command for testing
	exe := filepath.Join(t.TempDir(), "cronrange")
	build := exec.Command("go", "build", "-o", exe)
	if err := build.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	testTime := time.Date(2024, time.January, 2, 12, 30, 0, 0, time.UTC) // Tuesday, Jan 2, 2024 12:30 UTC
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		timeOffset time.Duration // offset from testTime
	}{
		{
			name:     "no arguments",
			args:     []string{},
			wantCode: 2,
		},
		{
			name:     "invalid expression format",
			args:     []string{"invalid"},
			wantCode: 2,
		},
		{
			name:     "matches range without command",
			args:     []string{"* * * *"},
			wantCode: 0,
		},
		{
			name:     "outside time range",
			args:     []string{"00:00-00:01 * * *"},
			wantCode: 1,
		},
		{
			name:       "outside day range",
			args:       []string{"* 1-5 * *"},
			timeOffset: time.Hour * 24 * 5, // Sunday
			wantCode:   1,
		},
		{
			name:     "valid command execution",
			args:     []string{"* * * *", "echo", "test"},
			wantCode: 0,
		},
		{
			name:     "command not found",
			args:     []string{"* * * *", "nonexistentcmd"},
			wantCode: 1,
		},
		{
			name:     "matches specific time",
			args:     []string{"12:00-13:00 * * *"},
			wantCode: 0,
		},
		{
			name:     "matches weekday",
			args:     []string{"* 1-5 * *"},
			wantCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(exe, tt.args...)

			// set test time
			testTimeWithOffset := testTime.Add(tt.timeOffset)
			cmd.Env = append(os.Environ(),
				"CRONRANGE_TEST_TIME="+testTimeWithOffset.Format(time.RFC3339))

			err := cmd.Run()
			var code int
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					code = exitErr.ExitCode()
				}
			}

			if code != tt.wantCode {
				t.Errorf("Expected exit code %d, got %d", tt.wantCode, code)
			}
		})
	}
}

func TestCommandOutput(t *testing.T) {
	// Build the command for testing
	exe := filepath.Join(t.TempDir(), "cronrange")
	build := exec.Command("go", "build", "-o", exe)
	if err := build.Run(); err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	t.Run("command output", func(t *testing.T) {
		cmd := exec.Command(exe, "* * * *", "echo", "test output")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if got := string(out); got != "test output\n" {
			t.Errorf("Expected output 'test output\\n', got %q", got)
		}
	})

	t.Run("command exit code", func(t *testing.T) {
		cmd := exec.Command(exe, "* * * *", "sh", "-c", "exit 42")
		err := cmd.Run()

		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 42 {
			t.Errorf("Expected exit code 42, got %v", err)
		}
	})
}
