# cronrange
[![Build Status](https://github.com/go-pkgz/cronrange/workflows/build/badge.svg)](https://github.com/go-pkgz/cronrange/actions) [![Coverage Status](https://coveralls.io/repos/github/go-pkgz/cronrange/badge.svg?branch=master)](https://coveralls.io/github/go-pkgz/cronrange?branch=master) [![Go Reference](https://pkg.go.dev/badge/github.com/go-pkgz/cronrange.svg)](https://pkg.go.dev/github.com/go-pkgz/cronrange)

`cronrange` is a Go package that provides a crontab-like format for expressing time ranges, particularly useful for defining recurring time windows. Unlike traditional crontab that defines specific moments in time, cronrange defines time periods when something should be active or eligible.

## Format

The format consists of four fields separated by whitespace:
```
time dow dom month
```

Where:
- `time`:  Time range in 24-hour format (HH:MM[:SS]-HH:MM[:SS]) or * for all day. Seconds are optional.
- `dow`:   Day of week (0-6, where 0=Sunday)
- `dom`:   Day of month (1-31)
- `month`: Month (1-12)

Multiple rules can be combined using semicolons (;).

Each field (except time) supports:
- Single values: "5"
- Lists:        "1,3,5"
- Ranges:       "1-5"
- Asterisk:     "*" for any/all values

## Examples

```
# Basic patterns
17:20-21:35 1-5 * *          # Weekdays from 5:20 PM to 9:35 PM
17:20:15-21:35:16 1-5 * *    # Weekdays from 5:20:15 PM to 9:35:16 PM
* 0,6 * *                    # All day on weekends
09:00-17:00 1-5 * 4-9        # Weekdays 9 AM to 5 PM, April through September
12:00-13:00 * 1,15 *         # Noon-1 PM on 1st and 15th of every month
23:00-07:00 * * *            # Overnight range from 11 PM to 7 AM, every day

# Multiple rules combined:
17:20-21:35 1-5 *;* * 0,6 * *              # Weekday evenings and all weekend
09:00-17:00 * *1-5; 10:00-16:00 * *6-12    # Different hours for different months
```

## Installation

```bash
go get github.com/go-pkgz/cronrange
```
## Library Usage

```go
import "github.com/go-pkgz/cronrange"

// Parse rules
rules, err := cronrange.Parse("17:20-21:35 1-5 *;* * 0,6 * *")
if err != nil {
    log.Fatal(err)
}

// Check if current time matches
if cronrange.Match(rules, time.Now()) {
    fmt.Println("Current time matches the rules")
}

// Check specific time
t := time.Date(2024, 1, 1, 18, 30, 0, 0, time.UTC)
if cronrange.Match(rules, t) {
    fmt.Println("Time matches the rules")
}

// Rules can be converted back to string format
fmt.Println(rules[0].String()) // "17:20-21:35 1-5 *"
```

Alternatively, you can use the `ParseFromReader` function to read rules from an `io.Reader`

## Error Handling

The package validates input and provides specific errors:
```go
// These will return errors
_, err1 := cronrange.Parse("25:00-26:00 1-5 * *")    // Invalid hours
_, err2 := cronrange.Parse("17:20-21:35 7 * *")      // Invalid day of week
_, err3 := cronrange.Parse("17:20-21:35 1-5 32 *")   // Invalid day of month
_, err4 := cronrange.Parse("17:20-21:35 1-5 * 13")   // Invalid month
_, err5 := cronrange.Parse("17:20-21:35 1-5 *")      // Wrong number of fields
```

## Command Line Utility

The package includes a command-line utility that can be used to execute commands within specified time ranges.
If the current time matches the range, the command is executed and the exit code indicates the result. If the time is outside the range, the command is not executed and the exit code is 1.

It can be run without the command just to check if the current time is within the range. in this case, the exit code will indicate the result, i.e. 0 if the time is within the range, 1 otherwise.

Install it with:

```bash
go install github.com/go-pkgz/cronrange/cmd@latest
```

### Usage

```bash
cronrange "TIME_RANGE" [command args...]
```

Examples:
```bash
# Check if current time is within range (exit code indicates result)
cronrange "17:20-21:35 1-5 * *"

# Execute command only if within range
cronrange "17:20-21:35 1-5 * *" echo "Running backup"
```

Exit codes:
- 0: Time matches range (or command executed successfully)
- 1: Time outside range (or command failed)
- 2: Invalid arguments or parsing error

