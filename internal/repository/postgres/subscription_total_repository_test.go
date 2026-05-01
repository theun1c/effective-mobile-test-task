package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	subscriptiontotal "github.com/theun1c/effective-mobile-test-task/internal/service/subscription_total"
)

func TestSubscriptionTotalRepositoryTotalCostCountsOnlyIntersectingSubscriptions(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	createRepo := NewSubscriptionRepository(db)
	totalRepo := NewSubscriptionTotalRepository(db)
	ctx := context.Background()
	userID := uuid.New()

	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Inside",
		Price:       100,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "04-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Starts Before",
		Price:       200,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "01-2025"),
		EndDate:     yearMonthPointer(t, "04-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Ends After",
		Price:       300,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "04-2025"),
		EndDate:     yearMonthPointer(t, "07-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Covers All",
		Price:       400,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "01-2025"),
		EndDate:     yearMonthPointer(t, "12-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "No End Date",
		Price:       500,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "04-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Outside",
		Price:       600,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "06-2025"),
		EndDate:     yearMonthPointer(t, "07-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "One Month",
		Price:       700,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "05-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})

	got, err := totalRepo.TotalCost(ctx, subscriptiontotal.Filter{
		Period: subscriptiontotal.Period{
			From: mustYearMonth(t, "03-2025"),
			To:   mustYearMonth(t, "05-2025"),
		},
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	const want = 4100
	if got != want {
		t.Fatalf("TotalCost() = %d, want %d", got, want)
	}
}

func TestSubscriptionTotalRepositoryTotalCostAppliesUserIDFilter(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	createRepo := NewSubscriptionRepository(db)
	totalRepo := NewSubscriptionTotalRepository(db)
	ctx := context.Background()

	targetUserID := uuid.New()
	otherUserID := uuid.New()

	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      targetUserID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       300,
		UserID:      otherUserID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Not Intersecting",
		Price:       500,
		UserID:      targetUserID,
		StartDate:   mustYearMonth(t, "07-2025"),
		EndDate:     yearMonthPointer(t, "08-2025"),
	})

	got, err := totalRepo.TotalCost(ctx, subscriptiontotal.Filter{
		Period: subscriptiontotal.Period{
			From: mustYearMonth(t, "03-2025"),
			To:   mustYearMonth(t, "05-2025"),
		},
		UserID: &targetUserID,
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	const want = 1200
	if got != want {
		t.Fatalf("TotalCost() = %d, want %d", got, want)
	}
}

func TestSubscriptionTotalRepositoryTotalCostAppliesServiceNameFilter(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	createRepo := NewSubscriptionRepository(db)
	totalRepo := NewSubscriptionTotalRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	serviceName := "Yandex Plus"

	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       400,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       300,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})

	got, err := totalRepo.TotalCost(ctx, subscriptiontotal.Filter{
		Period: subscriptiontotal.Period{
			From: mustYearMonth(t, "03-2025"),
			To:   mustYearMonth(t, "05-2025"),
		},
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	const want = 1200
	if got != want {
		t.Fatalf("TotalCost() = %d, want %d", got, want)
	}
}

func TestSubscriptionTotalRepositoryTotalCostAppliesCombinedFilters(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	createRepo := NewSubscriptionRepository(db)
	totalRepo := NewSubscriptionTotalRepository(db)
	ctx := context.Background()

	targetUserID := uuid.New()
	serviceName := "Yandex Plus"

	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       400,
		UserID:      targetUserID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       300,
		UserID:      targetUserID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})
	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: serviceName,
		Price:       500,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})

	got, err := totalRepo.TotalCost(ctx, subscriptiontotal.Filter{
		Period: subscriptiontotal.Period{
			From: mustYearMonth(t, "03-2025"),
			To:   mustYearMonth(t, "05-2025"),
		},
		UserID:      &targetUserID,
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	const want = 1200
	if got != want {
		t.Fatalf("TotalCost() = %d, want %d", got, want)
	}
}

func TestSubscriptionTotalRepositoryTotalCostReturnsZeroWhenNothingMatches(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	createRepo := NewSubscriptionRepository(db)
	totalRepo := NewSubscriptionTotalRepository(db)
	ctx := context.Background()
	userID := uuid.New()
	serviceName := "Nonexistent"

	createSubscriptionForTotal(t, createRepo, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   mustYearMonth(t, "03-2025"),
		EndDate:     yearMonthPointer(t, "05-2025"),
	})

	got, err := totalRepo.TotalCost(ctx, subscriptiontotal.Filter{
		Period: subscriptiontotal.Period{
			From: mustYearMonth(t, "06-2025"),
			To:   mustYearMonth(t, "08-2025"),
		},
		UserID:      &userID,
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	if got != 0 {
		t.Fatalf("TotalCost() = %d, want 0", got)
	}
}

func createSubscriptionForTotal(t *testing.T, repo *SubscriptionRepository, subscription domain.Subscription) {
	t.Helper()

	if _, err := repo.Create(context.Background(), subscription); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}
