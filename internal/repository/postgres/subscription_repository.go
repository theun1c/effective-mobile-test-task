package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

type SubscriptionRepository struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	now := time.Now().UTC()

	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO subscriptions (
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate.Time(),
		nullableDate(subscription.EndDate),
		now,
		now,
	)

	created, err := scanSubscription(row)
	if err != nil {
		return domain.Subscription{}, mapRepositoryError(err)
	}

	return created, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`,
		id,
	)

	subscription, err := scanSubscription(row)
	if err != nil {
		return domain.Subscription{}, mapRepositoryError(err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]domain.Subscription, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, mapRepositoryError(err)
	}
	defer rows.Close()

	subscriptions := make([]domain.Subscription, 0)
	for rows.Next() {
		subscription, scanErr := scanSubscription(rows)
		if scanErr != nil {
			return nil, mapRepositoryError(scanErr)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, mapRepositoryError(err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE subscriptions
		SET service_name = $2,
			price = $3,
			user_id = $4,
			start_date = $5,
			end_date = $6,
			updated_at = $7
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate.Time(),
		nullableDate(subscription.EndDate),
		time.Now().UTC(),
	)

	updated, err := scanSubscription(row)
	if err != nil {
		return domain.Subscription{}, mapRepositoryError(err)
	}

	return updated, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return mapRepositoryError(err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("determine affected rows: %w", err)
	}

	if affectedRows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanSubscription(scanner rowScanner) (domain.Subscription, error) {
	var (
		subscription domain.Subscription
		startDate    time.Time
		endDate      sql.NullTime
	)

	err := scanner.Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&startDate,
		&endDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)
	if err != nil {
		return domain.Subscription{}, err
	}

	subscription.StartDate, err = yearMonthFromTime(startDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	if endDate.Valid {
		parsedEndDate, parseErr := yearMonthFromTime(endDate.Time)
		if parseErr != nil {
			return domain.Subscription{}, parseErr
		}

		subscription.EndDate = &parsedEndDate
	}

	subscription.CreatedAt = subscription.CreatedAt.UTC()
	subscription.UpdatedAt = subscription.UpdatedAt.UTC()

	return subscription, nil
}

func nullableDate(value *yearmonth.YearMonth) any {
	if value == nil {
		return nil
	}

	return value.Time()
}

func yearMonthFromTime(value time.Time) (yearmonth.YearMonth, error) {
	parsed, err := yearmonth.Parse(value.UTC().Format("01-2006"))
	if err != nil {
		return yearmonth.YearMonth{}, fmt.Errorf("convert %s to year-month: %w", value, err)
	}

	return parsed, nil
}

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23514", "22P02", "22007":
			return fmt.Errorf("%w: %s", domain.ErrValidation, pgErr.Message)
		}
	}

	return err
}
