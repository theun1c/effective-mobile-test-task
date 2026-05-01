package yearmonth

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var strictYearMonthPattern = regexp.MustCompile(`^(0[1-9]|1[0-2])-(\d{4})$`)

type YearMonth struct {
	year  int
	month time.Month
}

func Parse(value string) (YearMonth, error) {
	matches := strictYearMonthPattern.FindStringSubmatch(value)
	if matches == nil {
		return YearMonth{}, fmt.Errorf("invalid year-month format %q: expected MM-YYYY", value)
	}

	month, err := strconv.Atoi(matches[1])
	if err != nil {
		return YearMonth{}, fmt.Errorf("parse month from %q: %w", value, err)
	}

	year, err := strconv.Atoi(matches[2])
	if err != nil {
		return YearMonth{}, fmt.Errorf("parse year from %q: %w", value, err)
	}

	return YearMonth{
		year:  year,
		month: time.Month(month),
	}, nil
}

func (ym YearMonth) String() string {
	return fmt.Sprintf("%02d-%04d", ym.month, ym.year)
}

func (ym YearMonth) Time() time.Time {
	return time.Date(ym.year, ym.month, 1, 0, 0, 0, 0, time.UTC)
}
