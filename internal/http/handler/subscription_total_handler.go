package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	httpresponse "github.com/theun1c/effective-mobile-test-task/internal/http/response"
	applogger "github.com/theun1c/effective-mobile-test-task/internal/logger"
)

type SubscriptionTotalService interface {
	TotalCost(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error)
}

type SubscriptionTotalHandler struct {
	service SubscriptionTotalService
	logger  *slog.Logger
}

func NewSubscriptionTotalHandler(service SubscriptionTotalService) *SubscriptionTotalHandler {
	return NewSubscriptionTotalHandlerWithLogger(service, applogger.Nop())
}

func NewSubscriptionTotalHandlerWithLogger(service SubscriptionTotalService, logger *slog.Logger) *SubscriptionTotalHandler {
	if logger == nil {
		logger = applogger.Nop()
	}

	return &SubscriptionTotalHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionTotalHandler) TotalCost(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeServiceError(h.logger, w, errors.New("subscription total service is not configured"), "subscription_total")
		return
	}

	req := dto.SubscriptionTotalQuery{
		From:        r.URL.Query().Get("from"),
		To:          r.URL.Query().Get("to"),
		UserID:      optionalQueryParam(r, "user_id"),
		ServiceName: optionalQueryParam(r, "service_name"),
	}

	h.logger.Info(
		"subscription total requested",
		"operation", "subscription_total",
		"from", req.From,
		"to", req.To,
		"user_id", optionalStringValue(req.UserID),
		"service_name", optionalStringValue(req.ServiceName),
	)

	response, err := h.service.TotalCost(r.Context(), req)
	if err != nil {
		writeServiceError(h.logger, w, err, "subscription_total")
		return
	}

	h.logger.Info(
		"subscription total calculated",
		"operation", "subscription_total",
		"total_cost", response.TotalCost,
	)

	httpresponse.JSON(w, http.StatusOK, response)
}

func optionalQueryParam(r *http.Request, name string) *string {
	values, ok := r.URL.Query()[name]
	if !ok {
		return nil
	}

	value := ""
	if len(values) > 0 {
		value = values[0]
	}

	return &value
}

func optionalStringValue(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}
