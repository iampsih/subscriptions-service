package util

import (
	"fmt"
	"time"
)

const MonthLayout = "01-2006" // MM-YYYY

func ParseMonth(s string) (time.Time, error) {
	t, err := time.ParseInLocation(MonthLayout, s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format, expected MM-YYYY: %w", err)
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FormatMonth(t time.Time) string {
	return t.In(time.UTC).Format(MonthLayout)
}
