package domain

import "errors"

var (
	ErrSubscriptionLimitExceeded = errors.New("subscription task limit exceeded")
	ErrNoActiveSubscription      = errors.New("no active subscription")
	ErrDuplicateAvitoID          = errors.New("duplicate avito_id")
)
