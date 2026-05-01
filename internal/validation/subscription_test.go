package validation

import (
	"errors"
	"testing"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
)

func TestValidateCreateSubscriptionRequestAcceptsValidPayload(t *testing.T) {
	req := dto.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     stringPointer("12-2025"),
	}

	if err := ValidateCreateSubscriptionRequest(req); err != nil {
		t.Fatalf("validate create request: %v", err)
	}
}

func TestValidateUpdateSubscriptionRequestAcceptsValidPayload(t *testing.T) {
	req := dto.UpdateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
	}

	if err := ValidateUpdateSubscriptionRequest(req); err != nil {
		t.Fatalf("validate update request: %v", err)
	}
}

func TestValidateCreateSubscriptionRequestRejectsInvalidPayload(t *testing.T) {
	req := dto.CreateSubscriptionRequest{
		ServiceName: "   ",
		Price:       0,
		UserID:      "invalid-uuid",
		StartDate:   "7-2025",
		EndDate:     stringPointer("06-2025"),
	}

	err := ValidateCreateSubscriptionRequest(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error kind, got %v", err)
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "service_name")
	assertHasFieldError(t, validationErr, "price")
	assertHasFieldError(t, validationErr, "user_id")
	assertHasFieldError(t, validationErr, "start_date")
}

func TestValidateCreateSubscriptionRequestRejectsInvalidEndDateFormat(t *testing.T) {
	req := dto.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     stringPointer("2025-07"),
	}

	err := ValidateCreateSubscriptionRequest(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "end_date")
}

func TestValidateCreateSubscriptionRequestRejectsEndDateBeforeStartDate(t *testing.T) {
	req := dto.CreateSubscriptionRequest{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
		StartDate:   "07-2025",
		EndDate:     stringPointer("06-2025"),
	}

	err := ValidateCreateSubscriptionRequest(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "end_date")
}

func assertHasFieldError(t *testing.T, validationErr *Error, field string) {
	t.Helper()

	for _, fieldErr := range validationErr.Fields {
		if fieldErr.Field == field {
			return
		}
	}

	t.Fatalf("expected field error for %q, got %+v", field, validationErr.Fields)
}

func stringPointer(value string) *string {
	return &value
}
