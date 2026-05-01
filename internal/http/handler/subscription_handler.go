package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	httpresponse "github.com/theun1c/effective-mobile-test-task/internal/http/response"
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
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSubscriptionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	subscription, err := h.service.Create(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusCreated, subscription)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(w, r)
	if !ok {
		return
	}

	subscription, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscription)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.service.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscriptions)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(w, r)
	if !ok {
		return
	}

	var req dto.UpdateSubscriptionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	subscription, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	httpresponse.JSON(w, http.StatusOK, subscription)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := subscriptionIDFromRequest(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func subscriptionIDFromRequest(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid subscription id", nil)
		return uuid.Nil, false
	}

	return id, true
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid JSON body", nil)
		return false
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpresponse.Error(w, http.StatusBadRequest, "invalid JSON body", nil)
		return false
	}

	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		httpresponse.Error(w, http.StatusBadRequest, domain.ErrValidation.Error(), validationDetails(err))
	case errors.Is(err, domain.ErrNotFound):
		httpresponse.Error(w, http.StatusNotFound, domain.ErrNotFound.Error(), nil)
	default:
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
