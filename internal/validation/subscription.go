package validation

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/dto"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Fields []FieldError `json:"errors"`
}

func (e *Error) Error() string {
	return domain.ErrValidation.Error()
}

func (e *Error) Is(target error) bool {
	return target == domain.ErrValidation
}

func (e *Error) add(field string, message string) {
	e.Fields = append(e.Fields, FieldError{
		Field:   field,
		Message: message,
	})
}

func (e *Error) empty() bool {
	return len(e.Fields) == 0
}

func ValidateCreateSubscriptionRequest(req dto.CreateSubscriptionRequest) error {
	return validateSubscriptionRequest(req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
}

func ValidateUpdateSubscriptionRequest(req dto.UpdateSubscriptionRequest) error {
	return validateSubscriptionRequest(req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
}

func validateSubscriptionRequest(serviceName string, price int, userID string, startDate string, endDate *string) error {
	validationErr := &Error{}

	if strings.TrimSpace(serviceName) == "" {
		validationErr.add("service_name", "must not be empty")
	}

	if price <= 0 {
		validationErr.add("price", "must be greater than 0")
	}

	if strings.TrimSpace(userID) == "" {
		validationErr.add("user_id", "must not be empty")
	} else if _, err := uuid.Parse(userID); err != nil {
		validationErr.add("user_id", "must be a valid UUID")
	}

	startYearMonth, startErr := parseRequiredYearMonth("start_date", startDate, validationErr)
	endYearMonth, endErr := parseOptionalYearMonth("end_date", endDate, validationErr)

	if startErr == nil && endErr == nil && endYearMonth != nil && endYearMonth.Time().Before(startYearMonth.Time()) {
		validationErr.add("end_date", "must not be earlier than start_date")
	}

	if validationErr.empty() {
		return nil
	}

	return validationErr
}

func parseRequiredYearMonth(field string, value string, validationErr *Error) (yearmonth.YearMonth, error) {
	if strings.TrimSpace(value) == "" {
		validationErr.add(field, "must not be empty")
		return yearmonth.YearMonth{}, errors.New("empty value")
	}

	parsed, err := yearmonth.Parse(value)
	if err != nil {
		validationErr.add(field, "must be in MM-YYYY format")
		return yearmonth.YearMonth{}, err
	}

	return parsed, nil
}

func parseOptionalYearMonth(field string, value *string, validationErr *Error) (*yearmonth.YearMonth, error) {
	if value == nil {
		return nil, nil
	}

	if strings.TrimSpace(*value) == "" {
		validationErr.add(field, "must be in MM-YYYY format")
		return nil, errors.New("empty value")
	}

	parsed, err := yearmonth.Parse(*value)
	if err != nil {
		validationErr.add(field, "must be in MM-YYYY format")
		return nil, err
	}

	return &parsed, nil
}
