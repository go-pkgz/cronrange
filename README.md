# cronrange

`cronrange` is a Go package that provides a crontab-like format for expressing time ranges, particularly useful for defining recurring time windows. Unlike traditional crontab that defines specific moments in time, cronrange defines time periods when something should be active or eligible.

## Format

The format consists of four fields separated by whitespace:

```
time dow dom month
```

Where:

- `time`:  Time range in 24-hour format (HH:MM-HH:MM) or * for all day
- `dow`:   Day of week (0-6, where 0=Sunday) 
- `dom`:   Day of month (1-31)
- `month`: Month (1-12)


Multiple rules can be combined using semicolons (;).

Each field (except time) supports:

- Single values: "5"
- Lists:        "1,3,5"
- Ranges:       "1-5"
-  Asterisk:     "*" for any/all values


## Examples

```
# Basic patterns
17:20-21:35 1-5 * *          # Weekdays from 5:20 PM to 9:35 PM
* 0,6 * *                    # All day on weekends
09:00-17:00 1-5 * 4-9        # Weekdays 9 AM to 5 PM, April through September
12:00-13:00 * 1,15 *         # Noon-1 PM on 1st and 15th of every month

# Multiple rules combined:
17:20-21:35 1-5 * *; * 0,6 * *    # Weekday evenings and all weekend
09:00-17:00 * * 1-5; 10:00-16:00 * * 6-12    # Different hours for different months
```

## Installation

```bash
go get github.com/go-pkgz/cronrange
```

## Usage

```go
import "github.com/go-pkgz/cronrange"

// Parse rules
rules, err := cronrange.Parse("17:20-21:35 1-5 * *; * 0,6 * *")
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
fmt.Println(rules[0].String()) // "17:20-21:35 1-5 * *"
```

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
