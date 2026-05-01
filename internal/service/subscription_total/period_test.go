package subscription_total

import (
	"testing"

	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

func TestIntersectSubscriptionPeriod(t *testing.T) {
	request := Period{
		From: mustYearMonth(t, "03-2025"),
		To:   mustYearMonth(t, "05-2025"),
	}

	testCases := []struct {
		name              string
		subscriptionStart string
		subscriptionEnd   *string
		wantFrom          string
		wantTo            string
		wantOK            bool
	}{
		{
			name:              "subscription fully inside request period",
			subscriptionStart: "03-2025",
			subscriptionEnd:   stringPointer("04-2025"),
			wantFrom:          "03-2025",
			wantTo:            "04-2025",
			wantOK:            true,
		},
		{
			name:              "subscription starts before request and ends inside",
			subscriptionStart: "01-2025",
			subscriptionEnd:   stringPointer("04-2025"),
			wantFrom:          "03-2025",
			wantTo:            "04-2025",
			wantOK:            true,
		},
		{
			name:              "subscription starts inside request and ends after",
			subscriptionStart: "04-2025",
			subscriptionEnd:   stringPointer("07-2025"),
			wantFrom:          "04-2025",
			wantTo:            "05-2025",
			wantOK:            true,
		},
		{
			name:              "subscription covers whole request period",
			subscriptionStart: "01-2025",
			subscriptionEnd:   stringPointer("12-2025"),
			wantFrom:          "03-2025",
			wantTo:            "05-2025",
			wantOK:            true,
		},
		{
			name:              "subscription without end date",
			subscriptionStart: "04-2025",
			subscriptionEnd:   nil,
			wantFrom:          "04-2025",
			wantTo:            "05-2025",
			wantOK:            true,
		},
		{
			name:              "subscription does not intersect request period",
			subscriptionStart: "06-2025",
			subscriptionEnd:   stringPointer("07-2025"),
			wantOK:            false,
		},
		{
			name:              "subscription active for exactly one month on boundary",
			subscriptionStart: "05-2025",
			subscriptionEnd:   stringPointer("05-2025"),
			wantFrom:          "05-2025",
			wantTo:            "05-2025",
			wantOK:            true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, ok := IntersectSubscriptionPeriod(
				mustYearMonth(t, testCase.subscriptionStart),
				mustOptionalYearMonth(t, testCase.subscriptionEnd),
				request,
			)

			if ok != testCase.wantOK {
				t.Fatalf("expected intersection=%v, got %v", testCase.wantOK, ok)
			}

			if !testCase.wantOK {
				return
			}

			if got.From.String() != testCase.wantFrom {
				t.Fatalf("expected from %q, got %q", testCase.wantFrom, got.From.String())
			}

			if got.To.String() != testCase.wantTo {
				t.Fatalf("expected to %q, got %q", testCase.wantTo, got.To.String())
			}
		})
	}
}

func TestActiveMonths(t *testing.T) {
	testCases := []struct {
		name              string
		subscriptionStart string
		subscriptionEnd   *string
		requestFrom       string
		requestTo         string
		wantMonths        int
	}{
		{
			name:              "subscription fully inside request period",
			subscriptionStart: "03-2025",
			subscriptionEnd:   stringPointer("04-2025"),
			requestFrom:       "01-2025",
			requestTo:         "12-2025",
			wantMonths:        2,
		},
		{
			name:              "subscription starts before request and ends inside",
			subscriptionStart: "01-2025",
			subscriptionEnd:   stringPointer("04-2025"),
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        2,
		},
		{
			name:              "subscription starts inside request and ends after",
			subscriptionStart: "04-2025",
			subscriptionEnd:   stringPointer("08-2025"),
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        2,
		},
		{
			name:              "subscription covers whole request period",
			subscriptionStart: "01-2025",
			subscriptionEnd:   stringPointer("12-2025"),
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        3,
		},
		{
			name:              "subscription without end date",
			subscriptionStart: "04-2025",
			subscriptionEnd:   nil,
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        2,
		},
		{
			name:              "subscription does not intersect request period",
			subscriptionStart: "06-2025",
			subscriptionEnd:   stringPointer("07-2025"),
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        0,
		},
		{
			name:              "subscription active for exactly one month",
			subscriptionStart: "05-2025",
			subscriptionEnd:   stringPointer("05-2025"),
			requestFrom:       "03-2025",
			requestTo:         "05-2025",
			wantMonths:        1,
		},
		{
			name:              "request period is one month",
			subscriptionStart: "05-2025",
			subscriptionEnd:   stringPointer("07-2025"),
			requestFrom:       "06-2025",
			requestTo:         "06-2025",
			wantMonths:        1,
		},
		{
			name:              "inclusive month boundaries count both ends",
			subscriptionStart: "03-2025",
			subscriptionEnd:   stringPointer("05-2025"),
			requestFrom:       "04-2025",
			requestTo:         "05-2025",
			wantMonths:        2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := Period{
				From: mustYearMonth(t, testCase.requestFrom),
				To:   mustYearMonth(t, testCase.requestTo),
			}

			got := ActiveMonths(
				mustYearMonth(t, testCase.subscriptionStart),
				mustOptionalYearMonth(t, testCase.subscriptionEnd),
				request,
			)

			if got != testCase.wantMonths {
				t.Fatalf("expected %d active months, got %d", testCase.wantMonths, got)
			}
		})
	}
}

func TestSubscriptionCostContribution(t *testing.T) {
	request := Period{
		From: mustYearMonth(t, "03-2025"),
		To:   mustYearMonth(t, "05-2025"),
	}

	got := SubscriptionCostContribution(
		400,
		mustYearMonth(t, "04-2025"),
		mustOptionalYearMonth(t, stringPointer("07-2025")),
		request,
	)

	if got != 800 {
		t.Fatalf("expected 800, got %d", got)
	}
}

func mustYearMonth(t *testing.T, value string) yearmonth.YearMonth {
	t.Helper()

	parsed, err := yearmonth.Parse(value)
	if err != nil {
		t.Fatalf("parse year-month %q: %v", value, err)
	}

	return parsed
}

func mustOptionalYearMonth(t *testing.T, value *string) *yearmonth.YearMonth {
	t.Helper()

	if value == nil {
		return nil
	}

	parsed := mustYearMonth(t, *value)
	return &parsed
}

func stringPointer(value string) *string {
	return &value
}
