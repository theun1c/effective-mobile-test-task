package handler

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	"github.com/theun1c/effective-mobile-test-task/internal/validation"
)

func TestSubscriptionTotalHandlerReturnsOK(t *testing.T) {
	service := &subscriptionTotalServiceStub{
		totalCostFunc: func(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
			if req.From != "03-2025" {
				t.Fatalf("TotalCost() From = %q, want 03-2025", req.From)
			}

			if req.To != "05-2025" {
				t.Fatalf("TotalCost() To = %q, want 05-2025", req.To)
			}

			if req.UserID == nil || *req.UserID != "60601fee-2bf1-4721-ae6f-7636e79a0cba" {
				t.Fatalf("TotalCost() UserID = %v, want 60601fee-2bf1-4721-ae6f-7636e79a0cba", req.UserID)
			}

			if req.ServiceName == nil || *req.ServiceName != "Yandex Plus" {
				t.Fatalf("TotalCost() ServiceName = %v, want Yandex Plus", req.ServiceName)
			}

			return dto.SubscriptionTotalResponse{TotalCost: 2400}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=03-2025&to=05-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex+Plus", nil)

	newSubscriptionTotalMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response dto.SubscriptionTotalResponse
	decodeJSON(t, recorder, &response)

	if response.TotalCost != 2400 {
		t.Fatalf("response total_cost = %d, want 2400", response.TotalCost)
	}
}

func TestSubscriptionTotalHandlerMapsValidationErrorToBadRequest(t *testing.T) {
	var logBuffer bytes.Buffer

	service := &subscriptionTotalServiceStub{
		totalCostFunc: func(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
			return dto.SubscriptionTotalResponse{}, &validation.Error{
				Fields: []validation.FieldError{
					{Field: "from", Message: "must be in MM-YYYY format"},
				},
			}
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=13-2025&to=05-2025", nil)

	newSubscriptionTotalMuxWithLogger(service, slog.New(slog.NewJSONHandler(&logBuffer, nil))).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	assertErrorResponse(t, recorder, "subscription validation failed", []validation.FieldError{
		{Field: "from", Message: "must be in MM-YYYY format"},
	})

	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, `"msg":"request rejected"`) {
		t.Fatalf("log output = %q, want request rejected message", logOutput)
	}

	if !strings.Contains(logOutput, `"reason":"validation_failed"`) {
		t.Fatalf("log output = %q, want validation_failed reason", logOutput)
	}
}

func TestSubscriptionTotalHandlerMapsUnexpectedErrorToInternalServerError(t *testing.T) {
	var logBuffer bytes.Buffer

	service := &subscriptionTotalServiceStub{
		totalCostFunc: func(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
			return dto.SubscriptionTotalResponse{}, errors.New("boom")
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=03-2025&to=05-2025", nil)

	newSubscriptionTotalMuxWithLogger(service, slog.New(slog.NewJSONHandler(&logBuffer, nil))).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	assertErrorResponse(t, recorder, "internal server error", nil)

	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, `"msg":"request failed"`) {
		t.Fatalf("log output = %q, want request failed message", logOutput)
	}

	if !strings.Contains(logOutput, `"operation":"subscription_total"`) {
		t.Fatalf("log output = %q, want operation field", logOutput)
	}
}

func TestSubscriptionTotalHandlerLogsRequestAndResult(t *testing.T) {
	var logBuffer bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	service := &subscriptionTotalServiceStub{
		totalCostFunc: func(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
			return dto.SubscriptionTotalResponse{TotalCost: 2400}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/total?from=03-2025&to=05-2025", nil)

	newSubscriptionTotalMuxWithLogger(service, logger).ServeHTTP(recorder, request)

	logOutput := logBuffer.String()
	if !strings.Contains(logOutput, `"msg":"subscription total requested"`) {
		t.Fatalf("log output = %q, want request message", logOutput)
	}

	if !strings.Contains(logOutput, `"msg":"subscription total calculated"`) {
		t.Fatalf("log output = %q, want calculated message", logOutput)
	}

	if !strings.Contains(logOutput, `"total_cost":2400`) {
		t.Fatalf("log output = %q, want total_cost field", logOutput)
	}
}

type subscriptionTotalServiceStub struct {
	totalCostCalls int
	totalCostFunc  func(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error)
}

func (s *subscriptionTotalServiceStub) TotalCost(ctx context.Context, req dto.SubscriptionTotalQuery) (dto.SubscriptionTotalResponse, error) {
	s.totalCostCalls++
	if s.totalCostFunc == nil {
		return dto.SubscriptionTotalResponse{}, nil
	}

	return s.totalCostFunc(ctx, req)
}

func newSubscriptionTotalMux(service *subscriptionTotalServiceStub) http.Handler {
	return newSubscriptionTotalMuxWithLogger(service, slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)))
}

func newSubscriptionTotalMuxWithLogger(service *subscriptionTotalServiceStub, logger *slog.Logger) http.Handler {
	handler := NewSubscriptionTotalHandlerWithLogger(service, logger)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /subscriptions/total", handler.TotalCost)
	return mux
}
