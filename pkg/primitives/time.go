package primitives

import (
	"errors"
	"time"
)

var (
	// ErrInvalidDuration indicates an invalid duration value
	ErrInvalidDuration = errors.New("invalid duration")
)

// Time wraps time.Time for temporal operations in the framework.
// Provides a consistent interface for time-based calculations.
type Time struct {
	value time.Time
}

// NewTime creates a Time from a time.Time value.
func NewTime(t time.Time) Time {
	return Time{value: t}
}

// Now returns the current time.
func Now() Time {
	return Time{value: time.Now()}
}

// Unix creates a Time from Unix timestamp (seconds since epoch).
func Unix(sec int64, nsec int64) Time {
	return Time{value: time.Unix(sec, nsec)}
}

// Add returns the time t+d.
func (t Time) Add(d Duration) Time {
	return Time{value: t.value.Add(d.value)}
}

// Sub returns the duration t-u.
func (t Time) Sub(u Time) Duration {
	return Duration{value: t.value.Sub(u.value)}
}

// Before reports whether the time instant t is before u.
func (t Time) Before(u Time) bool {
	return t.value.Before(u.value)
}

// After reports whether the time instant t is after u.
func (t Time) After(u Time) bool {
	return t.value.After(u.value)
}

// Equal reports whether t and u represent the same time instant.
func (t Time) Equal(u Time) bool {
	return t.value.Equal(u.value)
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC.
func (t Time) Unix() int64 {
	return t.value.Unix()
}

// UnixNano returns t as a Unix time, the number of nanoseconds elapsed
// since January 1, 1970 UTC.
func (t Time) UnixNano() int64 {
	return t.value.UnixNano()
}

// String returns the string representation of the Time.
func (t Time) String() string {
	return t.value.String()
}

// Format returns a textual representation of the time value formatted
// according to the layout defined by the argument.
func (t Time) Format(layout string) string {
	return t.value.Format(layout)
}

// Time returns the underlying time.Time value.
func (t Time) Time() time.Time {
	return t.value
}

// Duration wraps time.Duration for temporal durations in the framework.
type Duration struct {
	value time.Duration
}

// NewDuration creates a Duration from a time.Duration value.
func NewDuration(d time.Duration) Duration {
	return Duration{value: d}
}

// Seconds creates a Duration from seconds.
func Seconds(sec int64) Duration {
	return Duration{value: time.Duration(sec) * time.Second}
}

// Minutes creates a Duration from minutes.
func Minutes(min int64) Duration {
	return Duration{value: time.Duration(min) * time.Minute}
}

// Hours creates a Duration from hours.
func Hours(hr int64) Duration {
	return Duration{value: time.Duration(hr) * time.Hour}
}

// Days creates a Duration from days (24-hour periods).
func Days(days int64) Duration {
	return Duration{value: time.Duration(days) * 24 * time.Hour}
}

// Add returns the duration d+other.
func (d Duration) Add(other Duration) Duration {
	return Duration{value: d.value + other.value}
}

// Sub returns the duration d-other.
func (d Duration) Sub(other Duration) Duration {
	return Duration{value: d.value - other.value}
}

// Mul returns the duration d*factor.
func (d Duration) Mul(factor int64) Duration {
	return Duration{value: d.value * time.Duration(factor)}
}

// Div returns the duration d/divisor.
// Returns error if dividing by zero.
func (d Duration) Div(divisor int64) (Duration, error) {
	if divisor == 0 {
		return Duration{}, ErrDivisionByZero
	}
	return Duration{value: d.value / time.Duration(divisor)}, nil
}

// IsZero reports whether d represents the zero duration.
func (d Duration) IsZero() bool {
	return d.value == 0
}

// Abs returns the absolute value of d.
func (d Duration) Abs() Duration {
	if d.value < 0 {
		return Duration{value: -d.value}
	}
	return d
}

// Seconds returns the duration as a floating point number of seconds.
func (d Duration) Seconds() float64 {
	return d.value.Seconds()
}

// Minutes returns the duration as a floating point number of minutes.
func (d Duration) Minutes() float64 {
	return d.value.Minutes()
}

// Hours returns the duration as a floating point number of hours.
func (d Duration) Hours() float64 {
	return d.value.Hours()
}

// String returns the string representation of the Duration.
func (d Duration) String() string {
	return d.value.String()
}

// Duration returns the underlying time.Duration value.
func (d Duration) Duration() time.Duration {
	return d.value
}

// GreaterThan returns true if d > other.
func (d Duration) GreaterThan(other Duration) bool {
	return d.value > other.value
}

// LessThan returns true if d < other.
func (d Duration) LessThan(other Duration) bool {
	return d.value < other.value
}

// Equal returns true if d == other.
func (d Duration) Equal(other Duration) bool {
	return d.value == other.value
}
