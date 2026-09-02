package service

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/tokens"
)

// NewListingProcessor builds the configured listing_search processor.
func NewListingProcessor(pool *pgxpool.Pool, kind, avitoBase, rodUserDataDir, rodFetchMode string) (ListingProcessor, error) {
	switch kind {
	case tokens.ListingSearchProcessorStub, "":
		return NewStubListingProcessor(), nil
	case tokens.ListingSearchProcessorAvito:
		base := avitoBase
		if base == "" {
			base = tokens.AvitoHTTPSBase
		}
		return NewAvitoListingProcessor(pool, NewAvitoClient(WithAvitoHTTPBase(base))), nil
	case tokens.ListingSearchProcessorRod:
		fetch, err := newRodFetcher(avitoBase, rodUserDataDir, rodFetchMode)
		if err != nil {
			return nil, err
		}
		return NewRodAvitoListingProcessor(pool, fetch), nil
	default:
		return nil, fmt.Errorf("%s: %q", tokens.ErrMsgInvalidListingSearchProcessor, kind)
	}
}

func newRodFetcher(avitoBase, rodUserDataDir, rodFetchMode string) (AvitoPageFetcher, error) {
	mode := strings.TrimSpace(rodFetchMode)
	if mode == "" {
		mode = tokens.DefaultRodFetchMode
	}
	switch mode {
	case tokens.RodFetchModeHTTP:
		return NewHTTPPageFetcher(avitoBase), nil
	case tokens.RodFetchModeChrome:
		return NewRodPageFetcher(rodUserDataDir, avitoBase), nil
	default:
		return nil, fmt.Errorf("%s: %q", tokens.EnvRodFetchMode, mode)
	}
}
