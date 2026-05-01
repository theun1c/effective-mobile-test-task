package validation

import (
	"errors"
	"testing"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
)

func TestValidateSubscriptionTotalQueryAcceptsRequiredFieldsOnly(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From: "07-2025",
		To:   "12-2025",
	}

	if err := ValidateSubscriptionTotalQuery(req); err != nil {
		t.Fatalf("validate total query with required fields: %v", err)
	}
}

func TestValidateSubscriptionTotalQueryAcceptsOptionalFilters(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From:        "07-2025",
		To:          "12-2025",
		UserID:      stringPointer("60601fee-2bf1-4721-ae6f-7636e79a0cba"),
		ServiceName: stringPointer("Yandex Plus"),
	}

	if err := ValidateSubscriptionTotalQuery(req); err != nil {
		t.Fatalf("validate total query with optional filters: %v", err)
	}
}

func TestValidateSubscriptionTotalQueryRejectsMissingRequiredFields(t *testing.T) {
	req := dto.SubscriptionTotalQuery{}

	err := ValidateSubscriptionTotalQuery(req)
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

	assertHasFieldError(t, validationErr, "from")
	assertHasFieldError(t, validationErr, "to")
}

func TestValidateSubscriptionTotalQueryRejectsInvalidDateFormats(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From: "7-2025",
		To:   "2025-12",
	}

	err := ValidateSubscriptionTotalQuery(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "from")
	assertHasFieldError(t, validationErr, "to")
}

func TestValidateSubscriptionTotalQueryRejectsToEarlierThanFrom(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From: "08-2025",
		To:   "07-2025",
	}

	err := ValidateSubscriptionTotalQuery(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "to")
}

func TestValidateSubscriptionTotalQueryRejectsInvalidUserID(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From:   "07-2025",
		To:     "12-2025",
		UserID: stringPointer("invalid-uuid"),
	}

	err := ValidateSubscriptionTotalQuery(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "user_id")
}

func TestValidateSubscriptionTotalQueryRejectsEmptyServiceName(t *testing.T) {
	req := dto.SubscriptionTotalQuery{
		From:        "07-2025",
		To:          "12-2025",
		ServiceName: stringPointer("   "),
	}

	err := ValidateSubscriptionTotalQuery(req)
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	assertHasFieldError(t, validationErr, "service_name")
}
