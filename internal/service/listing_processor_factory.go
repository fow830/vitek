package service

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/tokens"
)

// NewListingProcessor builds the configured listing_search processor.
func NewListingProcessor(pool *pgxpool.Pool, kind, avitoBase, rodUserDataDir string) (ListingProcessor, error) {
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
		return NewRodAvitoListingProcessor(pool, NewRodPageFetcher(rodUserDataDir)), nil
	default:
		return nil, fmt.Errorf("%s: %q", tokens.ErrMsgInvalidListingSearchProcessor, kind)
	}
}
