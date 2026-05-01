package subscription

import (
	"time"

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
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
