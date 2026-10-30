package service

import (
	"errors"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// AvitoFetchFailure is a typed fetch outcome for heal / soft-pause policy.
type AvitoFetchFailure struct {
	State  string
	Status int
	cause  error
}

func newAvitoFetchFailure(state string, status int, soft bool) *AvitoFetchFailure {
	cause := domain.ErrListingSearchAvitoFetch
	if soft || tokens.AvitoFetchStateSoftFailure(state) {
		cause = domain.ErrListingSearchRateLimited
	}
	return &AvitoFetchFailure{State: state, Status: status, cause: cause}
}

func (e *AvitoFetchFailure) Error() string {
	if e == nil {
		return tokens.ErrMsgListingSearchAvitoFetch
	}
	if errors.Is(e.cause, domain.ErrListingSearchRateLimited) {
		return tokens.FormatAvitoFetchError(e.State, e.Status)
	}
	return tokens.FormatAvitoHardFetchError(e.State, e.Status)
}

func (e *AvitoFetchFailure) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Soft reports antibot / rate-limit failures that must not auto-PAUSE watches.
func (e *AvitoFetchFailure) Soft() bool {
	if e == nil {
		return false
	}
	return errors.Is(e.cause, domain.ErrListingSearchRateLimited)
}

func isAvitoSoftWatchFailure(err error) bool {
	if err == nil {
		return false
	}
	var fail *AvitoFetchFailure
	if errors.As(err, &fail) {
		return fail.Soft()
	}
	return errors.Is(err, domain.ErrListingSearchRateLimited)
}

// isAvitoFetchRecoverable is true when Chrome SERP may still succeed after HTTP heal fail.
func isAvitoFetchRecoverable(err error) bool {
	return errors.Is(err, domain.ErrListingSearchRateLimited) ||
		errors.Is(err, domain.ErrListingSearchAvitoFetch)
}

// NewAvitoFetchFailureForTest exposes typed failures to contracts (no string sniffing).
func NewAvitoFetchFailureForTest(state string, status int, soft bool) *AvitoFetchFailure {
	return newAvitoFetchFailure(state, status, soft)
}

func wrapAvitoGetFailure(last avitoGETResult) error {
	if last.state == "" && last.err != nil {
		return last.err
	}
	state := last.state
	if state == "" {
		state = tokens.AvitoFetchStateOther
	}
	soft := tokens.AvitoFetchStateHealable(state) || errors.Is(last.err, domain.ErrListingSearchRateLimited)
	if soft {
		return newAvitoFetchFailure(state, last.status, true)
	}
	if last.err != nil && !errors.Is(last.err, domain.ErrListingSearchAvitoFetch) && !errors.Is(last.err, domain.ErrListingSearchRateLimited) {
		return last.err
	}
	return newAvitoFetchFailure(state, last.status, false)
}

// errorFromAvitoFetchState maps classified state → typed soft/hard failure (nil if content).
func errorFromAvitoFetchState(state string, status int) error {
	if tokens.AvitoFetchStateIsContent(state) {
		return nil
	}
	return newAvitoFetchFailure(state, status, tokens.AvitoFetchStateSoftFailure(state))
}
