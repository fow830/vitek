package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/tokens"
)

func newAvitoJSONClient(base string) *AvitoClient {
	return NewAvitoClient(
		WithAvitoHTTPBase(base),
		WithProxyChangeIPURL(os.Getenv(tokens.EnvProxyChangeIPURL)),
	)
}

func normalizeAvitoHTTPBase(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return tokens.AvitoHTTPSBase
	}
	return base
}

// NewListingProcessor builds the configured listing_search processor.
func NewListingProcessor(pool *pgxpool.Pool, kind, avitoBase, rodUserDataDir, rodFetchMode string) (ListingProcessor, error) {
	base := normalizeAvitoHTTPBase(avitoBase)
	switch kind {
	case tokens.ListingSearchProcessorStub, "":
		return NewStubListingProcessor(), nil
	case tokens.ListingSearchProcessorAvito:
		return NewAvitoListingProcessor(pool, newAvitoJSONClient(base)), nil
	case tokens.ListingSearchProcessorRod:
		mode := strings.TrimSpace(rodFetchMode)
		if mode == "" {
			mode = tokens.DefaultRodFetchMode
		}
		chrome := NewRodPageFetcher(rodUserDataDir, base)
		switch mode {
		case tokens.RodFetchModeHTTP:
			// Default: HTTP rotate-until-clean, then Chrome SERP/warm.
			return NewRodAvitoListingProcessorWithClient(pool, chrome, newAvitoJSONClient(base)), nil
		case tokens.RodFetchModeChrome:
			// Chrome-only SERP/warm (no AvitoClient HTTP heal layer).
			return NewRodAvitoListingProcessorWithClient(pool, chrome, nil), nil
		default:
			return nil, fmt.Errorf("%s: %q", tokens.EnvRodFetchMode, mode)
		}
	default:
		return nil, fmt.Errorf("%s: %q", tokens.ErrMsgInvalidListingSearchProcessor, kind)
	}
}
