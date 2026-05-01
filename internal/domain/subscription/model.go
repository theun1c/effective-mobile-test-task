package subscription

import (
	"github.com/google/uuid"

	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

type Subscription struct {
	ID          uuid.UUID
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   yearmonth.YearMonth
	EndDate     *yearmonth.YearMonth
}
