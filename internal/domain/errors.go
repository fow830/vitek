package domain

import (
	"errors"

	"vitek/internal/tokens"
)

var (
	ErrSubscriptionLimitExceeded = errors.New(tokens.ErrMsgSubscriptionLimit)
	ErrNoActiveSubscription      = errors.New(tokens.ErrMsgNoActiveSubscription)
	ErrDuplicateAvitoID          = errors.New(tokens.ErrMsgDuplicateAvitoID)
	ErrServiceNotEntitled        = errors.New(tokens.ErrMsgServiceNotEntitled)
	ErrInvalidListingURL         = errors.New(tokens.ErrMsgInvalidListingURL)
)
