package yearmonth

import (
	"testing"
	"time"
)

func TestParseAcceptsStrictMMYYYY(t *testing.T) {
	value, err := Parse("07-2025")
	if err != nil {
		t.Fatalf("parse valid year-month: %v", err)
	}

	if got := value.String(); got != "07-2025" {
		t.Fatalf("expected original format, got %q", got)
	}
}

func TestParseRejectsInvalidFormats(t *testing.T) {
	testCases := []string{
		"7-2025",
		"13-2025",
		"00-2025",
		"2025-07",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			if _, err := Parse(testCase); err == nil {
				t.Fatalf("expected parse error for %q", testCase)
			}
		})
	}
}

func TestTimeReturnsFirstDayOfMonth(t *testing.T) {
	value, err := Parse("12-2026")
	if err != nil {
		t.Fatalf("parse valid year-month: %v", err)
	}

	got := value.Time()
	want := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
