package router

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/theun1c/effective-mobile-test-task/internal/dto"
)

func TestNewServesSwaggerUI(t *testing.T) {
	router := New(slog.New(slog.NewTextHandler(io.Discard, nil)), serviceStub{})

	request := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want html", contentType)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "/swagger/openapi.yaml") {
		t.Fatalf("body = %q, want link to raw openapi spec", body)
	}

	if strings.Contains(body, "jsdelivr") {
		t.Fatalf("body = %q, want self-contained swagger page without CDN assets", body)
	}
}

func TestNewServesSwaggerSpec(t *testing.T) {
	router := New(slog.New(slog.NewTextHandler(io.Discard, nil)), serviceStub{})

	request := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "yaml") && !strings.Contains(contentType, "text/plain") {
		t.Fatalf("Content-Type = %q, want yaml-like type", contentType)
	}

	if body := recorder.Body.String(); !strings.Contains(body, "openapi: 3.0.3") {
		t.Fatalf("body = %q, want openapi document", body)
	}
}

type serviceStub struct{}

func (serviceStub) Create(_ context.Context, _ dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	return dto.SubscriptionResponse{}, nil
}

func (serviceStub) GetByID(_ context.Context, _ uuid.UUID) (dto.SubscriptionResponse, error) {
	return dto.SubscriptionResponse{}, nil
}

func (serviceStub) List(_ context.Context) (dto.SubscriptionListResponse, error) {
	return dto.SubscriptionListResponse{}, nil
}

func (serviceStub) Update(_ context.Context, _ uuid.UUID, _ dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	return dto.SubscriptionResponse{}, nil
}

func (serviceStub) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}
