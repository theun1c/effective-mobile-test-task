package subscription_total

import (
	"time"

	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

type Period struct {
	From yearmonth.YearMonth
	To   yearmonth.YearMonth
}

func IntersectSubscriptionPeriod(
	subscriptionStart yearmonth.YearMonth,
	subscriptionEnd *yearmonth.YearMonth,
	request Period,
) (Period, bool) {
	intersectionStart := maxYearMonth(subscriptionStart, request.From)
	intersectionEnd := request.To

	if subscriptionEnd != nil {
		intersectionEnd = minYearMonth(*subscriptionEnd, request.To)
	}

	if intersectionEnd.Time().Before(intersectionStart.Time()) {
		return Period{}, false
	}

	return Period{
		From: intersectionStart,
		To:   intersectionEnd,
	}, true
}

func ActiveMonths(
	subscriptionStart yearmonth.YearMonth,
	subscriptionEnd *yearmonth.YearMonth,
	request Period,
) int {
	intersection, ok := IntersectSubscriptionPeriod(subscriptionStart, subscriptionEnd, request)
	if !ok {
		return 0
	}

	return monthsInclusive(intersection.From.Time(), intersection.To.Time())
}

func SubscriptionCostContribution(
	price int,
	subscriptionStart yearmonth.YearMonth,
	subscriptionEnd *yearmonth.YearMonth,
	request Period,
) int {
	return price * ActiveMonths(subscriptionStart, subscriptionEnd, request)
}

func maxYearMonth(left yearmonth.YearMonth, right yearmonth.YearMonth) yearmonth.YearMonth {
	if left.Time().Before(right.Time()) {
		return right
	}

	return left
}

func minYearMonth(left yearmonth.YearMonth, right yearmonth.YearMonth) yearmonth.YearMonth {
	if left.Time().Before(right.Time()) {
		return left
	}

	return right
}

func monthsInclusive(start time.Time, end time.Time) int {
	start = start.UTC()
	end = end.UTC()

	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
}
