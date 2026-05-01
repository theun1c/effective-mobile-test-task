package postgres

import (
	"context"
	"database/sql"

	subscriptiontotal "github.com/theun1c/effective-mobile-test-task/internal/service/subscription_total"
)

type SubscriptionTotalRepository struct {
	db *sql.DB
}

func NewSubscriptionTotalRepository(db *sql.DB) *SubscriptionTotalRepository {
	return &SubscriptionTotalRepository{db: db}
}

func (r *SubscriptionTotalRepository) TotalCost(ctx context.Context, filter subscriptiontotal.Filter) (int, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE start_date <= $1
			AND (end_date IS NULL OR end_date >= $2)
			AND ($3::uuid IS NULL OR user_id = $3)
			AND ($4::text IS NULL OR service_name = $4)`,
		filter.Period.To.Time(),
		filter.Period.From.Time(),
		filter.UserID,
		filter.ServiceName,
	)
	if err != nil {
		return 0, mapRepositoryError(err)
	}
	defer rows.Close()

	totalCost := 0
	for rows.Next() {
		subscription, scanErr := scanSubscription(rows)
		if scanErr != nil {
			return 0, mapRepositoryError(scanErr)
		}

		totalCost += subscriptiontotal.SubscriptionCostContribution(
			subscription.Price,
			subscription.StartDate,
			subscription.EndDate,
			filter.Period,
		)
	}

	if err := rows.Err(); err != nil {
		return 0, mapRepositoryError(err)
	}

	return totalCost, nil
}
