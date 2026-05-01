package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

func TestServiceCreateValidatesMapsAndReturnsResponse(t *testing.T) {
	repo := &repositoryStub{
		createFunc: func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
			if subscription.ID == uuid.Nil {
				t.Fatal("Create() passed nil ID to repository")
			}

			if subscription.ServiceName != "Yandex Plus" {
				t.Fatalf("Create() ServiceName = %q, want %q", subscription.ServiceName, "Yandex Plus")
			}

			if subscription.Price != 400 {
				t.Fatalf("Create() Price = %d, want 400", subscription.Price)
			}

			if subscription.UserID.String() != "60601fee-2bf1-4721-ae6f-7636e79a0cba" {
				t.Fatalf("Create() UserID = %s", subscription.UserID)
			}

			if subscription.StartDate.String() != "07-2025" {
				t.Fatalf("Create() StartDate = %s, want 07-2025", subscription.StartDate)
			}

			if subscription.EndDate == nil || subscription.EndDate.String() != "12-2025" {
				t.Fatalf("Create() EndDate = %v, want 12-2025", subscription.EndDate)
			}

			now := time.Date(2025, time.July, 1, 10, 0, 0, 0, time.UTC)
			subscription.CreatedAt = now
			subscription.UpdatedAt = now

			return subscription, nil
		},
	}

	service := New(repo)
	endDate := "12-2025"

	response, err := service.Create(context.Background(), dto.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     &endDate,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if repo.createCalls != 1 {
		t.Fatalf("Create() repository calls = %d, want 1", repo.createCalls)
	}

	if response.ServiceName != "Yandex Plus" {
		t.Fatalf("Create() response ServiceName = %q, want %q", response.ServiceName, "Yandex Plus")
	}

	if response.Price != 400 {
		t.Fatalf("Create() response Price = %d, want 400", response.Price)
	}

	if response.UserID != "60601fee-2bf1-4721-ae6f-7636e79a0cba" {
		t.Fatalf("Create() response UserID = %s", response.UserID)
	}

	if response.StartDate != "07-2025" {
		t.Fatalf("Create() response StartDate = %s, want 07-2025", response.StartDate)
	}

	if response.EndDate == nil || *response.EndDate != "12-2025" {
		t.Fatalf("Create() response EndDate = %v, want 12-2025", response.EndDate)
	}
}

func TestServiceCreateReturnsValidationErrorWithoutCallingRepository(t *testing.T) {
	repo := &repositoryStub{}
	service := New(repo)

	_, err := service.Create(context.Background(), dto.CreateSubscriptionRequest{
		ServiceName: "",
		Price:       0,
		UserID:      "invalid",
		StartDate:   "7-2025",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrValidation)
	}

	if repo.createCalls != 0 {
		t.Fatalf("Create() repository calls = %d, want 0", repo.createCalls)
	}
}

func TestServiceGetByIDReturnsMappedResponse(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	endDate := mustYearMonth(t, "09-2025")

	repo := &repositoryStub{
		getByIDFunc: func(ctx context.Context, gotID uuid.UUID) (domain.Subscription, error) {
			if gotID != id {
				t.Fatalf("GetByID() id = %s, want %s", gotID, id)
			}

			return domain.Subscription{
				ID:          id,
				ServiceName: "Netflix",
				Price:       999,
				UserID:      userID,
				StartDate:   mustYearMonth(t, "03-2025"),
				EndDate:     &endDate,
			}, nil
		},
	}

	service := New(repo)

	response, err := service.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if response.ID != id.String() {
		t.Fatalf("GetByID() response ID = %s, want %s", response.ID, id)
	}

	if response.UserID != userID.String() {
		t.Fatalf("GetByID() response UserID = %s, want %s", response.UserID, userID)
	}

	if response.StartDate != "03-2025" {
		t.Fatalf("GetByID() response StartDate = %s, want 03-2025", response.StartDate)
	}

	if response.EndDate == nil || *response.EndDate != "09-2025" {
		t.Fatalf("GetByID() response EndDate = %v, want 09-2025", response.EndDate)
	}
}

func TestServiceListReturnsMappedResponseList(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	firstUserID := uuid.New()
	secondUserID := uuid.New()

	repo := &repositoryStub{
		listFunc: func(ctx context.Context) ([]domain.Subscription, error) {
			return []domain.Subscription{
				{
					ID:          firstID,
					ServiceName: "Second",
					Price:       200,
					UserID:      firstUserID,
					StartDate:   mustYearMonth(t, "02-2025"),
				},
				{
					ID:          secondID,
					ServiceName: "First",
					Price:       100,
					UserID:      secondUserID,
					StartDate:   mustYearMonth(t, "01-2025"),
					EndDate:     yearMonthPointer(t, "04-2025"),
				},
			}, nil
		},
	}

	service := New(repo)

	response, err := service.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(response.Subscriptions) != 2 {
		t.Fatalf("List() len = %d, want 2", len(response.Subscriptions))
	}

	if response.Subscriptions[0].ID != firstID.String() || response.Subscriptions[1].ID != secondID.String() {
		t.Fatalf("List() IDs = [%s, %s]", response.Subscriptions[0].ID, response.Subscriptions[1].ID)
	}

	if response.Subscriptions[1].EndDate == nil || *response.Subscriptions[1].EndDate != "04-2025" {
		t.Fatalf("List() second EndDate = %v, want 04-2025", response.Subscriptions[1].EndDate)
	}
}

func TestServiceUpdateValidatesMapsAndReturnsResponse(t *testing.T) {
	id := uuid.New()
	userID := uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")

	repo := &repositoryStub{
		updateFunc: func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
			if subscription.ID != id {
				t.Fatalf("Update() ID = %s, want %s", subscription.ID, id)
			}

			if subscription.EndDate != nil {
				t.Fatalf("Update() EndDate = %v, want nil", subscription.EndDate)
			}

			return subscription, nil
		},
	}

	service := New(repo)

	response, err := service.Update(context.Background(), id, dto.UpdateSubscriptionRequest{
		ServiceName: "Updated Service",
		Price:       750,
		UserID:      userID.String(),
		StartDate:   "08-2025",
		EndDate:     nil,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if repo.updateCalls != 1 {
		t.Fatalf("Update() repository calls = %d, want 1", repo.updateCalls)
	}

	if response.ID != id.String() {
		t.Fatalf("Update() response ID = %s, want %s", response.ID, id)
	}

	if response.EndDate != nil {
		t.Fatalf("Update() response EndDate = %v, want nil", response.EndDate)
	}
}

func TestServiceDeletePropagatesNotFound(t *testing.T) {
	repo := &repositoryStub{
		deleteFunc: func(ctx context.Context, id uuid.UUID) error {
			return domain.ErrNotFound
		},
	}

	service := New(repo)

	err := service.Delete(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestServiceUpdateReturnsValidationErrorWithoutCallingRepository(t *testing.T) {
	repo := &repositoryStub{}
	service := New(repo)

	_, err := service.Update(context.Background(), uuid.New(), dto.UpdateSubscriptionRequest{
		ServiceName: "   ",
		Price:       -10,
		UserID:      "invalid",
		StartDate:   "13-2025",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrValidation)
	}

	if repo.updateCalls != 0 {
		t.Fatalf("Update() repository calls = %d, want 0", repo.updateCalls)
	}
}

type repositoryStub struct {
	createCalls int
	updateCalls int
	createFunc  func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	getByIDFunc func(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	listFunc    func(ctx context.Context) ([]domain.Subscription, error)
	updateFunc  func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	deleteFunc  func(ctx context.Context, id uuid.UUID) error
}

func (s *repositoryStub) Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	s.createCalls++
	if s.createFunc == nil {
		return domain.Subscription{}, nil
	}

	return s.createFunc(ctx, subscription)
}

func (s *repositoryStub) GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	if s.getByIDFunc == nil {
		return domain.Subscription{}, nil
	}

	return s.getByIDFunc(ctx, id)
}

func (s *repositoryStub) List(ctx context.Context) ([]domain.Subscription, error) {
	if s.listFunc == nil {
		return nil, nil
	}

	return s.listFunc(ctx)
}

func (s *repositoryStub) Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	s.updateCalls++
	if s.updateFunc == nil {
		return domain.Subscription{}, nil
	}

	return s.updateFunc(ctx, subscription)
}

func (s *repositoryStub) Delete(ctx context.Context, id uuid.UUID) error {
	if s.deleteFunc == nil {
		return nil
	}

	return s.deleteFunc(ctx, id)
}

func mustYearMonth(t *testing.T, value string) yearmonth.YearMonth {
	t.Helper()

	parsed, err := yearmonth.Parse(value)
	if err != nil {
		t.Fatalf("yearmonth.Parse(%q) error = %v", value, err)
	}

	return parsed
}

func yearMonthPointer(t *testing.T, value string) *yearmonth.YearMonth {
	t.Helper()

	parsed := mustYearMonth(t, value)
	return &parsed
}
