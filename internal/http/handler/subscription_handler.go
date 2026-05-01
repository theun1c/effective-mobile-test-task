package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	httpresponse "github.com/theun1c/effective-mobile-test-task/internal/http/response"
	applogger "github.com/theun1c/effective-mobile-test-task/internal/logger"
	"github.com/theun1c/effective-mobile-test-task/internal/validation"
)

type SubscriptionService interface {
	Create(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)
	List(ctx context.Context) (dto.SubscriptionListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SubscriptionHandler struct {
	service SubscriptionService
	logger  *slog.Logger
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return NewSubscriptionHandlerWithLogger(service, applogger.Nop())
}

func NewSubscriptionHandlerWithLogger(service SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	if logger == nil {
		logger = applogger.Nop()
	}

	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if !decodeJSONBody(h.logger, w, r, &req, "create_subscription") {
		return
	}

	subscription, err := h.service.Create(r.Context(), req)
	if err != nil {
		writeServiceError(h.logger, w, err, "create_subscription")
		return
	}

	httpresponse.JSON(w, http.StatusCreated, subscription)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(h.logger, w, r, "get_subscription")
	if !ok {
		return
	}

	subscription, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(h.logger, w, err, "get_subscription")
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscription)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(h.logger, w, err, "list_subscriptions")
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(h.logger, w, r, "update_subscription")
	if !ok {
		return
	}

	var req dto.UpdateSubscriptionRequest
	if !decodeJSONBody(h.logger, w, r, &req, "update_subscription") {
		return
	}

	subscription, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		writeServiceError(h.logger, w, err, "update_subscription")
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscription)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(h.logger, w, r, "delete_subscription")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(h.logger, w, err, "delete_subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func subscriptionIDFromRequest(logger *slog.Logger, w http.ResponseWriter, r *http.Request, operation string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		logger.Info(
			"request rejected",
			"operation", operation,
			"reason", "invalid_subscription_id",
			"path", r.URL.Path,
		)
		httpresponse.Error(w, http.StatusBadRequest, "invalid subscription id", nil)
		return uuid.Nil, false
	}

	return id, true
}

func decodeJSONBody(logger *slog.Logger, w http.ResponseWriter, r *http.Request, target any, operation string) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		logger.Info(
			"request rejected",
			"operation", operation,
			"reason", "invalid_json_body",
			"path", r.URL.Path,
			"error", err,
		)
		httpresponse.Error(w, http.StatusBadRequest, "invalid JSON body", nil)
		return false
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		logger.Info(
			"request rejected",
			"operation", operation,
			"reason", "invalid_json_body",
			"path", r.URL.Path,
		)
		httpresponse.Error(w, http.StatusBadRequest, "invalid JSON body", nil)
		return false
	}

	return true
}

func writeServiceError(logger *slog.Logger, w http.ResponseWriter, err error, operation string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		logger.Info(
			"request rejected",
			"operation", operation,
			"reason", "validation_failed",
			"error", err,
		)
		httpresponse.Error(w, http.StatusBadRequest, domain.ErrValidation.Error(), validationDetails(err))
	case errors.Is(err, domain.ErrNotFound):
		logger.Info(
			"request failed",
			"operation", operation,
			"reason", "not_found",
			"error", err,
		)
		httpresponse.Error(w, http.StatusNotFound, domain.ErrNotFound.Error(), nil)
	default:
		logger.Error(
			"request failed",
			"operation", operation,
			"error", err,
		)
		httpresponse.Error(w, http.StatusInternalServerError, "internal server error", nil)
	}
}

func validationDetails(err error) []httpresponse.ErrorDetail {
	var validationErr *validation.Error
	if !errors.As(err, &validationErr) {
		return nil
	}

	details := make([]httpresponse.ErrorDetail, 0, len(validationErr.Fields))
	for _, fieldErr := range validationErr.Fields {
		details = append(details, httpresponse.ErrorDetail{
			Field:   fieldErr.Field,
			Message: fieldErr.Message,
		})
	}

	return details
}
