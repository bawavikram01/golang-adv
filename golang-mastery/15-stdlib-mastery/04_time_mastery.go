//go:build ignore

// =============================================================================
// LESSON 15.4: time — Go's Time Package Deep Dive
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Time representation: time.Time internals (monotonic + wall clock)
// - The reference time: "Mon Jan 2 15:04:05 MST 2006" (why 1-2-3-4-5-6-7)
// - Formatting and parsing (every layout you'll ever need)
// - Duration, Ticker, Timer, After — concurrency-safe timing
// - Time zones: Location, LoadLocation, FixedZone
// - Comparison, truncation, rounding
// - Production patterns: timeouts, rate limiting, scheduling
// - Common pitfalls (== vs Equal, monotonic clock, zero time)
//
// THE KEY INSIGHT:
// Go's time.Time carries both a wall clock (for display) AND a monotonic
// clock (for measuring elapsed time). This means time.Since(start) is
// always accurate even if the system clock is adjusted. Most languages
// don't handle this correctly.
//
// RUN: go run 04_time_mastery.go
// =============================================================================

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== TIME MASTERY ===")
	fmt.Println()

	timeInternals()
	formattingParsing()
	durationMastery()
	tickerTimerPatterns()
	timeZones()
	comparisonArithmetic()
	commonPitfalls()
	productionPatterns()
}

// =============================================================================
// PART 1: Time Internals
// =============================================================================
func timeInternals() {
	fmt.Println("--- TIME INTERNALS ---")

	// time.Time contains:
	// 1. Wall clock: seconds since Jan 1, year 1, 00:00:00 UTC
	// 2. Monotonic clock: nanoseconds since process start
	// 3. Location (*time.Location): timezone information
	//
	// The monotonic clock is ONLY used for measuring duration.
	// It's immune to system clock adjustments (NTP, DST, manual changes).
	//
	// time.Now() returns BOTH wall + monotonic readings.
	// time.Date(), time.Parse() return ONLY wall clock (no monotonic).

	now := time.Now()
	fmt.Printf("  time.Now(): %v\n", now)
	// Notice the "m=+0.000123456" at the end — that's the monotonic reading!

	// Strip monotonic reading (for serialization, comparison)
	wallOnly := now.Round(0) // Round(0) strips monotonic clock
	fmt.Printf("  Wall only:  %v\n", wallOnly)

	// Zero value
	var zero time.Time
	fmt.Printf("  Zero time: %v, IsZero: %v\n", zero, zero.IsZero())
	// Zero is: January 1, year 1, 00:00:00 UTC
	// Use IsZero() to check for "not set" — NOT == time.Time{}

	// Unix timestamps
	fmt.Printf("  Unix seconds: %d\n", now.Unix())
	fmt.Printf("  Unix millis:  %d\n", now.UnixMilli())
	fmt.Printf("  Unix micros:  %d\n", now.UnixMicro())
	fmt.Printf("  Unix nanos:   %d\n", now.UnixNano())

	// From Unix timestamp
	fromUnix := time.Unix(1700000000, 0)
	fmt.Printf("  time.Unix(1700000000): %v\n", fromUnix)

	fmt.Println()
}

// =============================================================================
// PART 2: Formatting & Parsing — The Reference Time
// =============================================================================
func formattingParsing() {
	fmt.Println("--- FORMATTING & PARSING ---")

	// Go's time formatting uses a REFERENCE TIME instead of format codes.
	// The reference time is: Mon Jan 2 15:04:05 MST 2006
	//
	// WHY THOSE NUMBERS? They're sequential when written American-style:
	//   Month: 01 (January)
	//   Day:   02
	//   Hour:  03 (PM) or 15 (24-hour)
	//   Min:   04
	//   Sec:   05
	//   Year:  2006 (or 06)
	//   TZ:    -0700 or MST
	//
	// To make a custom format: write what the reference time looks like
	// in your desired format.

	now := time.Date(2024, 11, 15, 14, 30, 45, 123456789, time.UTC)

	// ─── Common formats ───
	fmt.Println("  Common formats:")
	fmt.Printf("    RFC3339:     %s\n", now.Format(time.RFC3339))     // 2024-11-15T14:30:45Z
	fmt.Printf("    RFC3339Nano: %s\n", now.Format(time.RFC3339Nano)) // 2024-11-15T14:30:45.123456789Z
	fmt.Printf("    Kitchen:     %s\n", now.Format(time.Kitchen))     // 2:30PM
	fmt.Printf("    DateTime:    %s\n", now.Format(time.DateTime))    // 2024-11-15 14:30:45 (Go 1.20+)
	fmt.Printf("    DateOnly:    %s\n", now.Format(time.DateOnly))    // 2024-11-15 (Go 1.20+)
	fmt.Printf("    TimeOnly:    %s\n", now.Format(time.TimeOnly))    // 14:30:45 (Go 1.20+)

	// ─── Custom formats ───
	fmt.Println("  Custom formats:")
	fmt.Printf("    YYYY/MM/DD:  %s\n", now.Format("2006/01/02"))              // 2024/11/15
	fmt.Printf("    DD-Mon-YYYY: %s\n", now.Format("02-Jan-2006"))             // 15-Nov-2024
	fmt.Printf("    12h:         %s\n", now.Format("3:04 PM"))                 // 2:30 PM
	fmt.Printf("    Full:        %s\n", now.Format("Monday, January 2, 2006")) // Friday, November 15, 2024
	fmt.Printf("    ISO week:    %s\n", now.Format("2006-W01"))                // week number
	fmt.Printf("    Millis:      %s\n", now.Format("15:04:05.000"))            // 14:30:45.123
	fmt.Printf("    Micros:      %s\n", now.Format("15:04:05.000000"))         // 14:30:45.123456

	// ─── Parsing ───
	fmt.Println("  Parsing:")

	t1, _ := time.Parse(time.RFC3339, "2024-11-15T14:30:45Z")
	fmt.Printf("    RFC3339: %v\n", t1)

	t2, _ := time.Parse("2006-01-02", "2024-11-15")
	fmt.Printf("    Date only: %v\n", t2)

	t3, _ := time.Parse("02/Jan/2006:15:04:05 -0700", "15/Nov/2024:14:30:45 +0530")
	fmt.Printf("    Nginx log: %v\n", t3)

	// Parse in a specific timezone
	loc, _ := time.LoadLocation("America/New_York")
	t4, _ := time.ParseInLocation("2006-01-02 15:04", "2024-11-15 14:30", loc)
	fmt.Printf("    ParseInLocation: %v\n", t4)

	// Parse error handling
	_, err := time.Parse(time.RFC3339, "not-a-date")
	fmt.Printf("    Parse error: %v\n", err)

	fmt.Println()
}

// =============================================================================
// PART 3: Duration
// =============================================================================
func durationMastery() {
	fmt.Println("--- DURATION MASTERY ---")

	// time.Duration is int64 nanoseconds.
	// Max duration: ~292 years. Smallest: 1 nanosecond.

	// ─── Creating durations ───
	fmt.Printf("  time.Second:      %v\n", time.Second)
	fmt.Printf("  time.Millisecond: %v\n", time.Millisecond)
	fmt.Printf("  5 * time.Minute:  %v\n", 5*time.Minute)
	fmt.Printf("  2.5 seconds:      %v\n", time.Duration(2.5*float64(time.Second)))

	// time.ParseDuration: parse human-readable durations
	d1, _ := time.ParseDuration("1h30m")
	d2, _ := time.ParseDuration("500ms")
	d3, _ := time.ParseDuration("2h45m30s")
	fmt.Printf("  ParseDuration: %v, %v, %v\n", d1, d2, d3)

	// ─── Duration methods ───
	d := 2*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Printf("  Duration: %v\n", d)
	fmt.Printf("    Hours:        %f\n", d.Hours())
	fmt.Printf("    Minutes:      %f\n", d.Minutes())
	fmt.Printf("    Seconds:      %f\n", d.Seconds())
	fmt.Printf("    Milliseconds: %d\n", d.Milliseconds())
	fmt.Printf("    Nanoseconds:  %d\n", d.Nanoseconds())
	fmt.Printf("    String:       %s\n", d.String())
	fmt.Printf("    Truncate(1m): %v\n", d.Truncate(time.Minute))
	fmt.Printf("    Round(1m):    %v\n", d.Round(time.Minute))

	// ─── Measuring elapsed time ───
	start := time.Now()
	// ... do work ...
	elapsed := time.Since(start) // = time.Now().Sub(start), uses monotonic clock
	fmt.Printf("  Elapsed: %v\n", elapsed)

	// time.Until: duration until a future time
	future := time.Now().Add(5 * time.Second)
	fmt.Printf("  Until(+5s): %v\n", time.Until(future))

	// ─── COMMON MISTAKE: integer * Duration ───
	// WRONG: time.Sleep(n * time.Second) where n is int → won't compile
	// RIGHT: time.Sleep(time.Duration(n) * time.Second)
	n := 3
	d4 := time.Duration(n) * time.Second
	fmt.Printf("  Dynamic duration: %v\n", d4)

	fmt.Println()
}

// =============================================================================
// PART 4: Ticker, Timer, After — Concurrency-Safe Timing
// =============================================================================
func tickerTimerPatterns() {
	fmt.Println("--- TICKER & TIMER PATTERNS ---")

	// ─── time.After: one-shot timeout (creates a channel) ───
	// WARNING: time.After LEAKS if the select exits before it fires!
	// Each call allocates a Timer that won't be GC'd until it fires.
	//
	// GOOD for one-time use:
	//   select {
	//   case result := <-ch:
	//       process(result)
	//   case <-time.After(5 * time.Second):
	//       return ErrTimeout
	//   }
	//
	// BAD in a loop: (leaks a Timer every iteration!)
	//   for {
	//       select {
	//       case msg := <-ch:
	//           process(msg)
	//       case <-time.After(1 * time.Second): // LEAK!
	//           checkHealth()
	//       }
	//   }

	// ─── time.NewTimer: reusable one-shot timer ───
	timer := time.NewTimer(100 * time.Millisecond)

	// Safe way to stop and drain:
	if !timer.Stop() {
		// Timer already fired, drain the channel
		select {
		case <-timer.C:
		default:
		}
	}
	// Reset for reuse (Go 1.23+ Reset is safe after Stop)
	timer.Reset(50 * time.Millisecond)
	<-timer.C
	fmt.Println("  Timer fired after 50ms")

	// ─── time.NewTicker: repeating interval ───
	ticker := time.NewTicker(25 * time.Millisecond)
	count := 0
	for range ticker.C {
		count++
		if count >= 3 {
			break
		}
	}
	ticker.Stop() // ALWAYS stop tickers when done (prevents goroutine leak)
	fmt.Printf("  Ticker: %d ticks\n", count)

	// ─── PRODUCTION PATTERN: Ticker in a goroutine ───
	// func startHealthCheck(ctx context.Context) {
	//     ticker := time.NewTicker(30 * time.Second)
	//     defer ticker.Stop()
	//     for {
	//         select {
	//         case <-ctx.Done():
	//             return
	//         case <-ticker.C:
	//             checkHealth()
	//         }
	//     }
	// }

	// ─── PRODUCTION PATTERN: Timer-based timeout in loop ───
	// timer := time.NewTimer(timeout)
	// defer timer.Stop()
	// for {
	//     if !timer.Stop() {
	//         select { case <-timer.C: default: }
	//     }
	//     timer.Reset(timeout)
	//     select {
	//     case msg := <-ch:
	//         process(msg)
	//     case <-timer.C:
	//         handleTimeout()
	//     }
	// }

	fmt.Println()
}

// =============================================================================
// PART 5: Time Zones
// =============================================================================
func timeZones() {
	fmt.Println("--- TIME ZONES ---")

	// time.Location represents a timezone.
	// time.UTC and time.Local are pre-defined.

	// ─── LoadLocation: from IANA timezone database ───
	ny, _ := time.LoadLocation("America/New_York")
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	india, _ := time.LoadLocation("Asia/Kolkata")

	now := time.Date(2024, 11, 15, 12, 0, 0, 0, time.UTC)

	fmt.Printf("  UTC:     %s\n", now.Format(time.RFC3339))
	fmt.Printf("  New York: %s\n", now.In(ny).Format(time.RFC3339))
	fmt.Printf("  Tokyo:   %s\n", now.In(tokyo).Format(time.RFC3339))
	fmt.Printf("  India:   %s\n", now.In(india).Format(time.RFC3339))

	// ─── FixedZone: for fixed-offset zones (when IANA DB not available) ───
	ist := time.FixedZone("IST", 5*3600+30*60) // +05:30
	fmt.Printf("  FixedZone IST: %s\n", now.In(ist).Format(time.RFC3339))

	// ─── IMPORTANT: Store in UTC, display in local ───
	// ALWAYS store times in UTC in your database.
	// Convert to user's timezone only for display.
	//
	// stored := time.Now().UTC()                    // store this
	// displayed := stored.In(userTimezone)           // display this

	// Zone() returns timezone name and offset
	name, offset := now.In(india).Zone()
	fmt.Printf("  Zone info: name=%q, offset=%d seconds\n", name, offset)

	fmt.Println()
}

// =============================================================================
// PART 6: Comparison & Arithmetic
// =============================================================================
func comparisonArithmetic() {
	fmt.Println("--- COMPARISON & ARITHMETIC ---")

	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)

	// ─── Comparison ───
	fmt.Printf("  Before: %v\n", t1.Before(t2)) // true
	fmt.Printf("  After:  %v\n", t1.After(t2))  // false
	fmt.Printf("  Equal:  %v\n", t1.Equal(t2))  // false

	// ─── Arithmetic ───
	// Add duration
	t3 := t1.Add(24 * time.Hour)
	fmt.Printf("  Add 24h: %s\n", t3.Format(time.DateOnly))

	// Subtract: get duration between two times
	diff := t2.Sub(t1)
	fmt.Printf("  Sub: %v (%.0f days)\n", diff, diff.Hours()/24)

	// AddDate: add years, months, days (handles month lengths correctly)
	t4 := t1.AddDate(1, 2, 3) // +1 year, +2 months, +3 days
	fmt.Printf("  AddDate(1,2,3): %s\n", t4.Format(time.DateOnly))

	// ─── Truncation & Rounding ───
	now := time.Date(2024, 11, 15, 14, 37, 42, 0, time.UTC)
	fmt.Printf("  Truncate(1h): %s\n", now.Truncate(time.Hour).Format(time.TimeOnly))       // 14:00:00
	fmt.Printf("  Round(1h):    %s\n", now.Round(time.Hour).Format(time.TimeOnly))          // 15:00:00
	fmt.Printf("  Truncate(15m): %s\n", now.Truncate(15*time.Minute).Format(time.TimeOnly)) // 14:30:00

	// ─── Extracting components ───
	fmt.Printf("  Year: %d, Month: %s, Day: %d\n", now.Year(), now.Month(), now.Day())
	fmt.Printf("  Hour: %d, Minute: %d, Second: %d\n", now.Hour(), now.Minute(), now.Second())
	fmt.Printf("  Weekday: %s\n", now.Weekday())
	y, w := now.ISOWeek()
	fmt.Printf("  ISO Week: %d-W%02d\n", y, w)
	fmt.Printf("  YearDay: %d\n", now.YearDay())

	fmt.Println()
}

// =============================================================================
// PART 7: Common Pitfalls
// =============================================================================
func commonPitfalls() {
	fmt.Println("--- COMMON PITFALLS ---")

	// ─── PITFALL 1: == vs Equal ───
	// == compares wall clock + location pointer + monotonic reading
	// Equal compares only the INSTANT in time (correct!)
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	loc, _ := time.LoadLocation("GMT") // GMT and UTC represent the same thing
	t2 := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)

	fmt.Printf("  ==:    %v (wrong! compares Location pointers)\n", t1 == t2)
	fmt.Printf("  Equal: %v (correct! compares instants)\n", t1.Equal(t2))
	// RULE: ALWAYS use .Equal() for time comparison

	// ─── PITFALL 2: Monotonic clock in serialization ───
	// time.Now() has monotonic reading. If you marshal and unmarshal,
	// the monotonic reading is lost. Sub() results will differ!
	now := time.Now()
	// Serialize and deserialize
	serialized := now.Format(time.RFC3339Nano)
	deserialized, _ := time.Parse(time.RFC3339Nano, serialized)
	// now has monotonic, deserialized doesn't
	// now.Sub(start) uses monotonic (accurate)
	// deserialized.Sub(start) uses wall clock (may drift)
	fmt.Printf("  Monotonic preserved after Parse: NO (%v lost)\n", now.Sub(deserialized))

	// ─── PITFALL 3: Zero time in JSON ───
	// time.Time zero value marshals to "0001-01-01T00:00:00Z"
	// Use *time.Time (pointer) + omitempty to omit when not set
	// Or use a custom marshaler that outputs null for zero time
	var zero time.Time
	fmt.Printf("  Zero time string: %q\n", zero.Format(time.RFC3339))

	// ─── PITFALL 4: time.Sleep in production ───
	// time.Sleep blocks the goroutine. In tests, it makes tests slow.
	// Use time.After/Timer/Ticker with select for cancelable waits.
	// In tests: use a clock interface or time.Now function that can be mocked.

	// ─── PITFALL 5: Month arithmetic edge cases ───
	// January 31 + 1 month = March 3 (February overflow!)
	jan31 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	feb := jan31.AddDate(0, 1, 0) // Expected Feb 31 → normalized to Mar 2
	fmt.Printf("  Jan 31 + 1 month = %s (not Feb!)\n", feb.Format(time.DateOnly))

	// ─── PITFALL 6: Comparing times across timezones ───
	// Two times in different zones that represent the same instant ARE Equal
	utcTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	ny, _ := time.LoadLocation("America/New_York")
	nyTime := utcTime.In(ny) // same instant, different display
	fmt.Printf("  UTC %s == NY %s: %v\n",
		utcTime.Format(time.RFC3339), nyTime.Format(time.RFC3339), utcTime.Equal(nyTime))

	fmt.Println()
}

// =============================================================================
// PART 8: Production Patterns
// =============================================================================
func productionPatterns() {
	fmt.Println("--- PRODUCTION PATTERNS ---")

	// ─── PATTERN 1: Mockable clock interface ───
	// Don't call time.Now() directly in business logic.
	// Inject a clock so tests can control time.
	//
	// type Clock interface {
	//     Now() time.Time
	//     Since(t time.Time) time.Duration
	//     NewTimer(d time.Duration) *time.Timer
	// }
	//
	// type realClock struct{}
	// func (realClock) Now() time.Time                         { return time.Now() }
	// func (realClock) Since(t time.Time) time.Duration        { return time.Since(t) }
	// func (realClock) NewTimer(d time.Duration) *time.Timer   { return time.NewTimer(d) }
	fmt.Println("  Pattern 1: Clock interface for testable code")

	// ─── PATTERN 2: Request timeout tracking ───
	// start := time.Now()
	// deadline := start.Add(30 * time.Second)
	// ctx, cancel := context.WithDeadline(ctx, deadline)
	// defer cancel()
	//
	// // Later, check remaining time:
	// remaining := time.Until(deadline)
	// if remaining < 5*time.Second {
	//     log.Warn("less than 5s remaining for request")
	// }
	fmt.Println("  Pattern 2: context.WithDeadline for request timeouts")

	// ─── PATTERN 3: Backoff timing ───
	// func backoff(attempt int) time.Duration {
	//     base := 100 * time.Millisecond
	//     max := 30 * time.Second
	//     d := base * time.Duration(1<<uint(attempt)) // exponential
	//     if d > max { d = max }
	//     // Add jitter: ±25%
	//     jitter := time.Duration(rand.Int63n(int64(d) / 2)) - d/4
	//     return d + jitter
	// }
	fmt.Println("  Pattern 3: Exponential backoff with jitter")

	// ─── PATTERN 4: Time-based caching ───
	// type CacheEntry[T any] struct {
	//     Value     T
	//     ExpiresAt time.Time
	// }
	//
	// func (e CacheEntry[T]) IsExpired() bool {
	//     return time.Now().After(e.ExpiresAt)
	// }
	fmt.Println("  Pattern 4: time.Now().After(expiry) for cache TTL")

	// ─── PATTERN 5: Debounce/throttle ───
	// var lastCall time.Time
	// func throttled(fn func()) {
	//     if time.Since(lastCall) < 100*time.Millisecond {
	//         return // too soon
	//     }
	//     lastCall = time.Now()
	//     fn()
	// }
	fmt.Println("  Pattern 5: time.Since for throttle/debounce")

	fmt.Println()
}
