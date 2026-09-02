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
	ErrTaskNotFound              = errors.New(tokens.ErrMsgTaskNotFound)
	ErrWatchNotFound             = errors.New(tokens.ErrMsgWatchNotFound)
	ErrForbidden                 = errors.New(tokens.ErrMsgForbidden)
	ErrListingSearchNoProxy      = errors.New(tokens.ErrMsgListingSearchNoProxy)
	ErrListingSearchNoAccount    = errors.New(tokens.ErrMsgListingSearchNoAccount)
	ErrListingSearchAvitoFetch   = errors.New(tokens.ErrMsgListingSearchAvitoFetch)
	ErrListingSearchRodFilterOnly = errors.New(tokens.ErrMsgListingSearchRodFilterOnly)
	ErrResourceNotFound          = errors.New(tokens.ErrMsgResourceNotFound)
)
