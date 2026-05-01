package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	"github.com/theun1c/effective-mobile-test-task/internal/validation"
)

func TestSubscriptionHandlerCreateReturnsCreatedResponse(t *testing.T) {
	service := &serviceStub{
		createFunc: func(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error) {
			if req.ServiceName != "Yandex Plus" {
				t.Fatalf("Create() ServiceName = %q, want %q", req.ServiceName, "Yandex Plus")
			}

			if req.StartDate != "07-2025" {
				t.Fatalf("Create() StartDate = %q, want %q", req.StartDate, "07-2025")
			}

			return dto.SubscriptionResponse{
				ID:          "d65a5e7a-0d26-4514-b4ac-8fe95583f07f",
				ServiceName: req.ServiceName,
				Price:       req.Price,
				UserID:      req.UserID,
				StartDate:   req.StartDate,
				EndDate:     req.EndDate,
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions",
		strings.NewReader(`{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}`),
	)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var response dto.SubscriptionResponse
	decodeJSON(t, recorder, &response)

	if response.ID != "d65a5e7a-0d26-4514-b4ac-8fe95583f07f" {
		t.Fatalf("response ID = %q", response.ID)
	}
}

func TestSubscriptionHandlerCreateReturnsBadRequestForInvalidJSON(t *testing.T) {
	service := &serviceStub{}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(`{"service_name":`))

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	if service.createCalls != 0 {
		t.Fatalf("Create() service calls = %d, want 0", service.createCalls)
	}

	assertErrorResponse(t, recorder, "invalid JSON body", nil)
}

func TestSubscriptionHandlerCreateMapsValidationErrorToBadRequest(t *testing.T) {
	service := &serviceStub{
		createFunc: func(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error) {
			return dto.SubscriptionResponse{}, &validation.Error{
				Fields: []validation.FieldError{
					{Field: "user_id", Message: "must be a valid UUID"},
				},
			}
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/subscriptions",
		strings.NewReader(`{"service_name":"Yandex Plus","price":400,"user_id":"invalid","start_date":"07-2025"}`),
	)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	assertErrorResponse(t, recorder, domain.ErrValidation.Error(), []validation.FieldError{
		{Field: "user_id", Message: "must be a valid UUID"},
	})
}

func TestSubscriptionHandlerListReturnsOK(t *testing.T) {
	service := &serviceStub{
		listFunc: func(ctx context.Context) (dto.SubscriptionListResponse, error) {
			return dto.SubscriptionListResponse{
				Subscriptions: []dto.SubscriptionResponse{
					{
						ID:          "d65a5e7a-0d26-4514-b4ac-8fe95583f07f",
						ServiceName: "Netflix",
						Price:       999,
						UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
						StartDate:   "07-2025",
					},
				},
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response dto.SubscriptionListResponse
	decodeJSON(t, recorder, &response)

	if len(response.Subscriptions) != 1 {
		t.Fatalf("subscriptions len = %d, want 1", len(response.Subscriptions))
	}
}

func TestSubscriptionHandlerGetByIDReturnsOK(t *testing.T) {
	id := uuid.New()
	service := &serviceStub{
		getByIDFunc: func(ctx context.Context, gotID uuid.UUID) (dto.SubscriptionResponse, error) {
			if gotID != id {
				t.Fatalf("GetByID() id = %s, want %s", gotID, id)
			}

			return dto.SubscriptionResponse{
				ID:          id.String(),
				ServiceName: "Spotify",
				Price:       300,
				UserID:      uuid.New().String(),
				StartDate:   "06-2025",
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/"+id.String(), nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response dto.SubscriptionResponse
	decodeJSON(t, recorder, &response)

	if response.ID != id.String() {
		t.Fatalf("response ID = %s, want %s", response.ID, id)
	}
}

func TestSubscriptionHandlerGetByIDReturnsBadRequestForInvalidUUID(t *testing.T) {
	service := &serviceStub{}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/not-a-uuid", nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	if service.getByIDCalls != 0 {
		t.Fatalf("GetByID() service calls = %d, want 0", service.getByIDCalls)
	}

	assertErrorResponse(t, recorder, "invalid subscription id", nil)
}

func TestSubscriptionHandlerGetByIDMapsNotFoundTo404(t *testing.T) {
	id := uuid.New()
	service := &serviceStub{
		getByIDFunc: func(ctx context.Context, gotID uuid.UUID) (dto.SubscriptionResponse, error) {
			return dto.SubscriptionResponse{}, domain.ErrNotFound
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions/"+id.String(), nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	assertErrorResponse(t, recorder, domain.ErrNotFound.Error(), nil)
}

func TestSubscriptionHandlerUpdateReturnsOK(t *testing.T) {
	id := uuid.New()
	service := &serviceStub{
		updateFunc: func(ctx context.Context, gotID uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
			if gotID != id {
				t.Fatalf("Update() id = %s, want %s", gotID, id)
			}

			return dto.SubscriptionResponse{
				ID:          id.String(),
				ServiceName: req.ServiceName,
				Price:       req.Price,
				UserID:      req.UserID,
				StartDate:   req.StartDate,
				EndDate:     req.EndDate,
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/subscriptions/"+id.String(),
		strings.NewReader(`{"service_name":"Yandex Plus","price":500,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025","end_date":"12-2025"}`),
	)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response dto.SubscriptionResponse
	decodeJSON(t, recorder, &response)

	if response.EndDate == nil || *response.EndDate != "12-2025" {
		t.Fatalf("response EndDate = %v, want 12-2025", response.EndDate)
	}
}

func TestSubscriptionHandlerDeleteReturnsNoContent(t *testing.T) {
	id := uuid.New()
	service := &serviceStub{
		deleteFunc: func(ctx context.Context, gotID uuid.UUID) error {
			if gotID != id {
				t.Fatalf("Delete() id = %s, want %s", gotID, id)
			}

			return nil
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/subscriptions/"+id.String(), nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if body := recorder.Body.String(); body != "" {
		t.Fatalf("body = %q, want empty", body)
	}
}

func TestSubscriptionHandlerMapsUnexpectedErrorToInternalServerError(t *testing.T) {
	service := &serviceStub{
		listFunc: func(ctx context.Context) (dto.SubscriptionListResponse, error) {
			return dto.SubscriptionListResponse{}, errors.New("boom")
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)

	newSubscriptionMux(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	assertErrorResponse(t, recorder, "internal server error", nil)
}

type serviceStub struct {
	createCalls  int
	getByIDCalls int
	updateCalls  int
	deleteCalls  int
	createFunc   func(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error)
	getByIDFunc  func(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)
	listFunc     func(ctx context.Context) (dto.SubscriptionListResponse, error)
	updateFunc   func(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (s *serviceStub) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	s.createCalls++
	if s.createFunc == nil {
		return dto.SubscriptionResponse{}, nil
	}

	return s.createFunc(ctx, req)
}

func (s *serviceStub) GetByID(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	s.getByIDCalls++
	if s.getByIDFunc == nil {
		return dto.SubscriptionResponse{}, nil
	}

	return s.getByIDFunc(ctx, id)
}

func (s *serviceStub) List(ctx context.Context) (dto.SubscriptionListResponse, error) {
	if s.listFunc == nil {
		return dto.SubscriptionListResponse{}, nil
	}

	return s.listFunc(ctx)
}

func (s *serviceStub) Update(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	s.updateCalls++
	if s.updateFunc == nil {
		return dto.SubscriptionResponse{}, nil
	}

	return s.updateFunc(ctx, id, req)
}

func (s *serviceStub) Delete(ctx context.Context, id uuid.UUID) error {
	s.deleteCalls++
	if s.deleteFunc == nil {
		return nil
	}

	return s.deleteFunc(ctx, id)
}

type errorResponse struct {
	Error  string                  `json:"error"`
	Errors []validation.FieldError `json:"errors,omitempty"`
}

func newSubscriptionMux(service *serviceStub) http.Handler {
	handler := NewSubscriptionHandler(service)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /subscriptions", handler.Create)
	mux.HandleFunc("GET /subscriptions", handler.List)
	mux.HandleFunc("GET /subscriptions/{id}", handler.GetByID)
	mux.HandleFunc("PUT /subscriptions/{id}", handler.Update)
	mux.HandleFunc("DELETE /subscriptions/{id}", handler.Delete)
	return mux
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantError string, wantFields []validation.FieldError) {
	t.Helper()

	var response errorResponse
	decodeJSON(t, recorder, &response)

	if response.Error != wantError {
		t.Fatalf("error = %q, want %q", response.Error, wantError)
	}

	if len(response.Errors) != len(wantFields) {
		t.Fatalf("errors len = %d, want %d", len(response.Errors), len(wantFields))
	}

	for i := range wantFields {
		if response.Errors[i] != wantFields[i] {
			t.Fatalf("errors[%d] = %+v, want %+v", i, response.Errors[i], wantFields[i])
		}
	}
}
