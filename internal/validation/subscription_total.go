package validation

import (
	"strings"

	"github.com/google/uuid"

	"github.com/theun1c/effective-mobile-test-task/internal/dto"
)

func ValidateSubscriptionTotalQuery(req dto.SubscriptionTotalQuery) error {
	validationErr := &Error{}

	fromYearMonth, fromErr := parseRequiredYearMonth("from", req.From, validationErr)
	toYearMonth, toErr := parseRequiredYearMonth("to", req.To, validationErr)

	if fromErr == nil && toErr == nil && toYearMonth.Time().Before(fromYearMonth.Time()) {
		validationErr.add("to", "must not be earlier than from")
	}

	if req.UserID != nil {
		if _, err := uuid.Parse(strings.TrimSpace(*req.UserID)); err != nil {
			validationErr.add("user_id", "must be a valid UUID")
		}
	}

	if req.ServiceName != nil && strings.TrimSpace(*req.ServiceName) == "" {
		validationErr.add("service_name", "must not be empty")
	}

	if validationErr.empty() {
		return nil
	}

	return validationErr
}
