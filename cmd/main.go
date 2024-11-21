// cmd/main.go
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-pkgz/cronrange"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s TIME_RANGE [command args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s \"17:20-21:35 1-5 * *\" echo hello\n", os.Args[0])
		os.Exit(2)
	}

	// parse cronrange expression
	rules, err := cronrange.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cronrange: %v\n", err)
		os.Exit(2)
	}

	// get current time or use test time if provided
	now := time.Now()
	if testTime := os.Getenv("CRONRANGE_TEST_TIME"); testTime != "" {
		parsed, err := time.Parse(time.RFC3339, testTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing test time: %v\n", err)
			os.Exit(2)
		}
		now = parsed
	}

	// check if current time matches the rules
	if !cronrange.Match(rules, now) {
		os.Exit(1)
	}

	// if no command provided, just exit with success
	if len(os.Args) == 2 {
		os.Exit(0)
	}

	// execute the command
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
