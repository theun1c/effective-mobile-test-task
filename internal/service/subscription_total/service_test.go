package subscription_total

import (
	"context"
	"errors"
	"testing"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
)

func TestServiceTotalCostValidatesMapsAndReturnsResponse(t *testing.T) {
	repo := &repositoryStub{
		totalCostFunc: func(ctx context.Context, filter Filter) (int, error) {
			if filter.Period.From.String() != "03-2025" {
				t.Fatalf("TotalCost() filter from = %s, want 03-2025", filter.Period.From)
			}

			if filter.Period.To.String() != "05-2025" {
				t.Fatalf("TotalCost() filter to = %s, want 05-2025", filter.Period.To)
			}

			if filter.UserID == nil || filter.UserID.String() != "60601fee-2bf1-4721-ae6f-7636e79a0cba" {
				t.Fatalf("TotalCost() filter user_id = %v, want 60601fee-2bf1-4721-ae6f-7636e79a0cba", filter.UserID)
			}

			if filter.ServiceName == nil || *filter.ServiceName != "Yandex Plus" {
				t.Fatalf("TotalCost() filter service_name = %v, want Yandex Plus", filter.ServiceName)
			}

			return 2400, nil
		},
	}

	service := New(repo)
	userID := "60601fee-2bf1-4721-ae6f-7636e79a0cba"
	serviceName := "Yandex Plus"

	response, err := service.TotalCost(context.Background(), dto.SubscriptionTotalQuery{
		From:        "03-2025",
		To:          "05-2025",
		UserID:      &userID,
		ServiceName: &serviceName,
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	if repo.totalCostCalls != 1 {
		t.Fatalf("TotalCost() repository calls = %d, want 1", repo.totalCostCalls)
	}

	if response.TotalCost != 2400 {
		t.Fatalf("TotalCost() response total_cost = %d, want 2400", response.TotalCost)
	}
}

func TestServiceTotalCostPassesNilOptionalFilters(t *testing.T) {
	repo := &repositoryStub{
		totalCostFunc: func(ctx context.Context, filter Filter) (int, error) {
			if filter.UserID != nil {
				t.Fatalf("TotalCost() filter user_id = %v, want nil", filter.UserID)
			}

			if filter.ServiceName != nil {
				t.Fatalf("TotalCost() filter service_name = %v, want nil", filter.ServiceName)
			}

			return 1000, nil
		},
	}

	service := New(repo)

	response, err := service.TotalCost(context.Background(), dto.SubscriptionTotalQuery{
		From: "04-2025",
		To:   "04-2025",
	})
	if err != nil {
		t.Fatalf("TotalCost() error = %v", err)
	}

	if response.TotalCost != 1000 {
		t.Fatalf("TotalCost() response total_cost = %d, want 1000", response.TotalCost)
	}
}

func TestServiceTotalCostReturnsValidationErrorWithoutCallingRepository(t *testing.T) {
	repo := &repositoryStub{}
	service := New(repo)
	userID := "invalid"

	_, err := service.TotalCost(context.Background(), dto.SubscriptionTotalQuery{
		From:   "13-2025",
		To:     "02-2025",
		UserID: &userID,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("TotalCost() error = %v, want %v", err, domain.ErrValidation)
	}

	if repo.totalCostCalls != 0 {
		t.Fatalf("TotalCost() repository calls = %d, want 0", repo.totalCostCalls)
	}
}

func TestServiceTotalCostPropagatesRepositoryError(t *testing.T) {
	expectedErr := errors.New("boom")
	repo := &repositoryStub{
		totalCostFunc: func(ctx context.Context, filter Filter) (int, error) {
			return 0, expectedErr
		},
	}

	service := New(repo)

	_, err := service.TotalCost(context.Background(), dto.SubscriptionTotalQuery{
		From: "03-2025",
		To:   "05-2025",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("TotalCost() error = %v, want %v", err, expectedErr)
	}

	if repo.totalCostCalls != 1 {
		t.Fatalf("TotalCost() repository calls = %d, want 1", repo.totalCostCalls)
	}
}

type repositoryStub struct {
	totalCostCalls int
	totalCostFunc  func(ctx context.Context, filter Filter) (int, error)
}

func (s *repositoryStub) TotalCost(ctx context.Context, filter Filter) (int, error) {
	s.totalCostCalls++
	if s.totalCostFunc == nil {
		return 0, nil
	}

	return s.totalCostFunc(ctx, filter)
}
