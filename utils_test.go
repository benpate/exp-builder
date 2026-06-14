package builder

import (
	"testing"
	"time"

	"github.com/benpate/exp"
	"github.com/stretchr/testify/require"
)

// TestParseValue_Operators exercises the "OP:value" parsing across the full set
// of recognized operator aliases, plus the fall-through cases.
func TestParseValue_Operators(t *testing.T) {

	run := func(input string, expectedOp string, expectedValue string) {
		op, value := parseValue(input, exp.OperatorEqual)
		require.Equal(t, expectedOp, op, "input=%q", input)
		require.Equal(t, expectedValue, value, "input=%q", input)
	}

	// Recognized operator prefixes are standardized.
	run("EQ:John", "=", "John")
	run("NE:John", "!=", "John")
	run("GT:7", ">", "7")
	run("GTE:7", ">=", "7")
	run("LT:7", "<", "7")
	run("LTE:7", "<=", "7")
	run("CONTAINS:abc", "CONTAINS", "abc")
	run("BEGINS:abc", "BEGINS", "abc")

	// Operator matching is case-insensitive.
	run("gt:7", ">", "7")

	// No colon: the whole input is the value, with the default operator.
	run("plain value", "=", "plain value")

	// Unrecognized operator before the colon: the default operator is used and
	// the entire input (including the colon) becomes the value.
	run("bogus:value", "=", "bogus:value")

	// Only the FIRST colon splits operator from value.
	run("EQ:a:b:c", "=", "a:b:c")

	// Empty input returns empty operator and value.
	run("", "", "")
}

func TestSliceNotEmpty(t *testing.T) {
	require.False(t, sliceNotEmpty(nil))
	require.False(t, sliceNotEmpty([]string{}))
	require.False(t, sliceNotEmpty([]string{""}))
	require.False(t, sliceNotEmpty([]string{"", ""}))
	require.True(t, sliceNotEmpty([]string{"value"}))
	require.True(t, sliceNotEmpty([]string{"", "value"}))
}

// midnightUTC returns today at 00:00:00 UTC, matching how parseTimeRange anchors
// its relative ranges.
func midnightUTC() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

// TestParseTimeRange_DayRanges verifies the day-counted ranges exactly.
func TestParseTimeRange_DayRanges(t *testing.T) {

	today := midnightUTC()

	run := func(keyword string, expectedBegin time.Time, expectedEnd time.Time) {
		begin, end := parseTimeRange(keyword)
		require.Equal(t, expectedBegin, begin, keyword)
		require.Equal(t, expectedEnd, end, keyword)
		require.True(t, begin.Before(end), keyword)
	}

	run("past-30-days", today.AddDate(0, 0, -30), today)
	run("past-60-days", today.AddDate(0, 0, -60), today)
	run("past-90-days", today.AddDate(0, 0, -90), today)
	run("past-180-days", today.AddDate(0, 0, -180), today)
	run("past-365-days", today.AddDate(0, 0, -365), today)

	run("next-30-days", today, today.AddDate(0, 0, 30))
	run("next-60-days", today, today.AddDate(0, 0, 60))
	run("next-90-days", today, today.AddDate(0, 0, 90))
	run("next-180-days", today, today.AddDate(0, 0, 180))
	run("next-365-days", today, today.AddDate(0, 0, 365))

	run("yesterday", today.AddDate(0, 0, -1), today)
	run("today", today, today.AddDate(0, 0, 1))
	run("tomorrow", today.AddDate(0, 0, 1), today.AddDate(0, 0, 2))
}

// TestParseTimeRange_CalendarRanges verifies the calendar-aligned ranges with
// invariants (rather than re-deriving the exact dates, which would just mirror
// the implementation): each range is non-zero, ordered, and starts at UTC midnight.
func TestParseTimeRange_CalendarRanges(t *testing.T) {

	keywords := []string{
		"this-week", "next-week",
		"last-month", "this-month", "next-month",
		"last-year", "this-year", "next-year",
	}

	for _, keyword := range keywords {
		begin, end := parseTimeRange(keyword)

		require.False(t, begin.IsZero(), keyword)
		require.False(t, end.IsZero(), keyword)
		require.True(t, begin.Before(end), keyword)

		// Ranges are anchored to midnight UTC.
		require.Equal(t, time.UTC, begin.Location(), keyword)
		require.Equal(t, 0, begin.Hour(), keyword)
		require.Equal(t, 0, begin.Minute(), keyword)
		require.Equal(t, 0, begin.Second(), keyword)
	}
}

func TestParseTimeRange_Unknown(t *testing.T) {
	// An unrecognized keyword returns two zero times.
	begin, end := parseTimeRange("not-a-real-range")
	require.True(t, begin.IsZero())
	require.True(t, end.IsZero())
}

// FuzzParseValue confirms the operator/value parser never panics and that its
// output invariants hold for arbitrary input.
func FuzzParseValue(f *testing.F) {

	seeds := []string{"", "John", "EQ:John", "gt:7", "bogus:value", "::", ":x", "a:b:c"}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {

		operator, value := parseValue(input, exp.OperatorEqual)

		// Empty input yields empty operator and value.
		if input == "" {
			require.Equal(t, "", operator)
			require.Equal(t, "", value)
			return
		}

		// The value is always derived from the input, never longer than it.
		require.LessOrEqual(t, len(value), len(input))

		// A non-default operator must be one that exp recognizes as standard.
		if operator != exp.OperatorEqual {
			_, ok := exp.OperatorOk(operator)
			require.True(t, ok, "operator=%q", operator)
		}
	})
}

// FuzzParseTimeRange confirms the time-range parser never panics and always
// returns either a zero pair (unrecognized) or a correctly ordered range.
func FuzzParseTimeRange(f *testing.F) {

	seeds := []string{"today", "yesterday", "past-30-days", "next-year", "this-week", "", "garbage"}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {

		begin, end := parseTimeRange(value)

		// Unrecognized keywords return two zero times.
		if begin.IsZero() && end.IsZero() {
			return
		}

		// Any recognized range is non-empty and correctly ordered.
		require.True(t, begin.Before(end), "value=%q", value)
	})
}
