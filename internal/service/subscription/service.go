package subscription

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	applogger "github.com/theun1c/effective-mobile-test-task/internal/logger"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
	"github.com/theun1c/effective-mobile-test-task/internal/validation"
)

type Repository interface {
	Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	List(ctx context.Context) ([]domain.Subscription, error)
	Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	repository Repository
	logger     *slog.Logger
}

func New(repository Repository) *Service {
	return NewWithLogger(repository, applogger.Nop())
}

func NewWithLogger(repository Repository, logger *slog.Logger) *Service {
	if logger == nil {
		logger = applogger.Nop()
	}

	return &Service{
		repository: repository,
		logger:     logger,
	}
}

func (s *Service) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	if err := validation.ValidateCreateSubscriptionRequest(req); err != nil {
		return dto.SubscriptionResponse{}, err
	}

	subscription, err := newDomainSubscription(uuid.New(), req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	created, err := s.repository.Create(ctx, subscription)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	s.logger.Info(
		"subscription created",
		"subscription_id", created.ID,
		"user_id", created.UserID,
		"service_name", created.ServiceName,
	)

	return toSubscriptionResponse(created), nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	subscription, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return toSubscriptionResponse(subscription), nil
}

func (s *Service) List(ctx context.Context) (dto.SubscriptionListResponse, error) {
	subscriptions, err := s.repository.List(ctx)
	if err != nil {
		return dto.SubscriptionListResponse{}, err
	}

	response := dto.SubscriptionListResponse{
		Subscriptions: make([]dto.SubscriptionResponse, 0, len(subscriptions)),
	}

	for _, subscription := range subscriptions {
		response.Subscriptions = append(response.Subscriptions, toSubscriptionResponse(subscription))
	}

	return response, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	if err := validation.ValidateUpdateSubscriptionRequest(req); err != nil {
		return dto.SubscriptionResponse{}, err
	}

	subscription, err := newDomainSubscription(id, req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	updated, err := s.repository.Update(ctx, subscription)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	s.logger.Info(
		"subscription updated",
		"subscription_id", updated.ID,
		"user_id", updated.UserID,
		"service_name", updated.ServiceName,
	)

	return toSubscriptionResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("subscription deleted", "subscription_id", id)

	return nil
}

func newDomainSubscription(
	id uuid.UUID,
	serviceName string,
	price int,
	userID string,
	startDate string,
	endDate *string,
) (domain.Subscription, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("%w: parse user_id: %v", domain.ErrValidation, err)
	}

	parsedStartDate, err := yearmonth.Parse(startDate)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("%w: parse start_date: %v", domain.ErrValidation, err)
	}

	parsedEndDate, err := parseOptionalYearMonth(endDate)
	if err != nil {
		return domain.Subscription{}, err
	}

	return domain.Subscription{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      parsedUserID,
		StartDate:   parsedStartDate,
		EndDate:     parsedEndDate,
	}, nil
}

func parseOptionalYearMonth(value *string) (*yearmonth.YearMonth, error) {
	if value == nil {
		return nil, nil
	}

	parsed, err := yearmonth.Parse(*value)
	if err != nil {
		return nil, fmt.Errorf("%w: parse end_date: %v", domain.ErrValidation, err)
	}

	return &parsed, nil
}

func toSubscriptionResponse(subscription domain.Subscription) dto.SubscriptionResponse {
	return dto.SubscriptionResponse{
		ID:          subscription.ID.String(),
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID.String(),
		StartDate:   subscription.StartDate.String(),
		EndDate:     toOptionalString(subscription.EndDate),
	}
}

func toOptionalString(value *yearmonth.YearMonth) *string {
	if value == nil {
		return nil
	}

	formatted := value.String()
	return &formatted
}
