package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// AvitoListingProcessor uses proxy + Avito web JSON to find similar listings.
type AvitoListingProcessor struct {
	proxies *Proxies
	avito   *AvitoAccounts
	client  *AvitoClient
}

func NewAvitoListingProcessor(pool *pgxpool.Pool, client *AvitoClient) *AvitoListingProcessor {
	if client == nil {
		client = NewAvitoClient()
	}
	return &AvitoListingProcessor{
		proxies: NewProxies(pool),
		avito:   NewAvitoAccounts(pool),
		client:  client,
	}
}

func (p *AvitoListingProcessor) FindSimilar(ctx context.Context, listingURL string) ([]SimilarListing, error) {
	if !tokens.ValidListingURL(listingURL) {
		return nil, domain.ErrInvalidListingURL
	}
	proxies, err := p.proxies.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, domain.ErrListingSearchNoProxy
	}
	if _, err := p.avito.PickActive(ctx); err != nil {
		return nil, domain.ErrListingSearchNoAccount
	}
	proxy := proxies[0].Endpoint
	if tokens.IsListingFilterURL(listingURL) {
		return p.client.FindFromFilterURL(ctx, proxy, listingURL)
	}
	itemID := tokens.ListingIDFromURL(listingURL)
	return p.client.FindSimilar(ctx, proxy, itemID)
}
