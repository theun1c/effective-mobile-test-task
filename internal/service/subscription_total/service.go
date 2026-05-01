package subscription_total

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
	"github.com/theun1c/effective-mobile-test-task/internal/validation"
)

type Filter struct {
	Period      Period
	UserID      *uuid.UUID
	ServiceName *string
}

type Repository interface {
	TotalCost(ctx context.Context, filter Filter) (int, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) TotalCost(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
	if err := validation.ValidateSubscriptionTotalQuery(req); err != nil {
		return dto.SubscriptionTotalResponse{}, err
	}

	filter, err := newFilter(req)
	if err != nil {
		return dto.SubscriptionTotalResponse{}, err
	}

	totalCost, err := s.repository.TotalCost(ctx, filter)
	if err != nil {
		return dto.SubscriptionTotalResponse{}, err
	}

	return dto.SubscriptionTotalResponse{
		TotalCost: totalCost,
	}, nil
}

func newFilter(req dto.SubscriptionTotalQuery) (Filter, error) {
	from, err := yearmonth.Parse(req.From)
	if err != nil {
		return Filter{}, fmt.Errorf("%w: parse from: %v", domain.ErrValidation, err)
	}

	to, err := yearmonth.Parse(req.To)
	if err != nil {
		return Filter{}, fmt.Errorf("%w: parse to: %v", domain.ErrValidation, err)
	}

	userID, err := parseOptionalUUID(req.UserID)
	if err != nil {
		return Filter{}, err
	}

	return Filter{
		Period: Period{
			From: from,
			To:   to,
		},
		UserID:      userID,
		ServiceName: req.ServiceName,
	}, nil
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}

	parsed, err := uuid.Parse(strings.TrimSpace(*value))
	if err != nil {
		return nil, fmt.Errorf("%w: parse user_id: %v", domain.ErrValidation, err)
	}

	return &parsed, nil
}
