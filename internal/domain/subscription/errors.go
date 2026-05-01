package subscription

import "errors"

var (
	ErrNotFound   = errors.New("subscription not found")
	ErrValidation = errors.New("subscription validation failed")
)
