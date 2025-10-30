package builder

import (
	"strings"
	"time"

	"github.com/benpate/exp"
)

func parseTimeRange(value string) (time.Time, time.Time) {

	switch value {

	case "past-30-days":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		beginDate := endDate.AddDate(0, 0, -30)
		return beginDate, endDate

	case "past-60-days":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		beginDate := endDate.AddDate(0, 0, -60)
		return beginDate, endDate

	case "past-90-days":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		beginDate := endDate.AddDate(0, 0, -90)
		return beginDate, endDate

	case "past-180-days":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		beginDate := endDate.AddDate(0, 0, -180)
		return beginDate, endDate

	case "past-365-days":
		now := time.Now()
		endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		beginDate := endDate.AddDate(0, 0, -365)
		return beginDate, endDate

	case "next-30-days":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 30)
		return beginDate, endDate

	case "next-60-days":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 60)
		return beginDate, endDate

	case "next-90-days":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 90)
		return beginDate, endDate

	case "next-180-days":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 180)
		return beginDate, endDate

	case "next-365-days":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 365)
		return beginDate, endDate

	case "yesterday":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
		endDate := beginDate.AddDate(0, 0, 1)
		return beginDate, endDate

	case "today":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 0, 1)
		return beginDate, endDate

	case "tomorrow":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		endDate := beginDate.AddDate(0, 0, 1)
		return beginDate, endDate

	case "this-week":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -int(now.Weekday()))
		endDate := beginDate.AddDate(0, 0, 7)
		return beginDate, endDate

	case "next-week":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 7-int(now.Weekday()))
		endDate := beginDate.AddDate(0, 0, 7)
		return beginDate, endDate

	case "last-month":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
		endDate := beginDate.AddDate(0, 1, 0)
		return beginDate, endDate

	case "this-month":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(0, 1, 0)
		return beginDate, endDate

	case "next-month":
		now := time.Now()
		beginDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
		endDate := beginDate.AddDate(0, 1, 0)
		return beginDate, endDate

	case "last-year":
		now := time.Now()
		beginDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC).AddDate(-1, 0, 0)
		endDate := beginDate.AddDate(1, 0, 0)
		return beginDate, endDate

	case "this-year":
		now := time.Now()
		beginDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := beginDate.AddDate(1, 0, 0)
		return beginDate, endDate

	case "next-year":
		now := time.Now()
		beginDate := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC).AddDate(1, 0, 0)
		endDate := beginDate.AddDate(1, 0, 0)
		return beginDate, endDate
	}

	return time.Time{}, time.Time{}
}

// parseValue parses a single string into an operator and a value, using the form
// "OP:Value" -- so the string "EQ:John" will return ("=", "John").
// If a valid operator is defined in the input, then it is returned along
// with the remaining string as the criteria value.
// Otherwise, the defaultOperator is used.
func parseValue(input string, defaultOperator string) (string, string) {

	if len(input) > 0 {

		// If the input contains a colon, then split it into OPERATOR and VALUE
		if operator, value, found := strings.Cut(input, ":"); found {

			if operator, ok := exp.OperatorOk(operator); ok {
				return operator, value
			}
		}

		// Otherwise, use the default operator argument
		return defaultOperator, input
	}

	return "", ""
}

func sliceNotEmpty(slice []string) bool {

	for _, value := range slice {
		if len(value) > 0 {
			return true
		}
	}

	return false
}

// isOdd returns TRUE if a number is odd
func isOdd(value int) bool {
	return value%2 > 0
}
