package service

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/repository"
	"vitek/internal/tokens"
)

// AvitoPageFetcher loads an Avito page HTML (browser session).
type AvitoPageFetcher interface {
	FetchHTML(ctx context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error)
}

// ProfileWarmer runs a multi-step sticky-profile warm (home → dwell → secondary).
// WarmSession prefers this over a single FetchHTML probe when available.
type ProfileWarmer interface {
	WarmProfile(ctx context.Context, proxyEndpoint, userDataDir string) (string, error)
}

// RodAvitoListingProcessor finds filter SERP listings via optional AvitoClient (HTTP heal)
// then Chrome/HTML page fetcher.
type RodAvitoListingProcessor struct {
	bindings *Bindings
	fetch    AvitoPageFetcher
	client   *AvitoClient

	mu       sync.Mutex
	lastMeta tokens.ListingFilterMeta
}

func NewRodAvitoListingProcessor(pool *pgxpool.Pool, fetch AvitoPageFetcher) *RodAvitoListingProcessor {
	return NewRodAvitoListingProcessorWithClient(pool, fetch, nil)
}

// NewRodAvitoListingProcessorWithClient prefers healed HTTP SERP when client is set.
func NewRodAvitoListingProcessorWithClient(pool *pgxpool.Pool, fetch AvitoPageFetcher, client *AvitoClient) *RodAvitoListingProcessor {
	if fetch == nil {
		fetch = NewRodPageFetcher("", "")
	}
	return &RodAvitoListingProcessor{
		bindings: NewBindings(pool),
		fetch:    fetch,
		client:   client,
	}
}

// HasHTTPHeal reports AvitoClient HTTP rotate-until-clean before Chrome SERP.
func (p *RodAvitoListingProcessor) HasHTTPHeal() bool {
	return p != nil && p.client != nil
}

// PageFetcher exposes the session-warm / HTML SERP fetcher (Chrome when configured).
func (p *RodAvitoListingProcessor) PageFetcher() AvitoPageFetcher {
	if p == nil {
		return nil
	}
	return p.fetch
}

// LastFilterMeta returns meta enriched from the last successful SERP HTML fetch.
func (p *RodAvitoListingProcessor) LastFilterMeta() tokens.ListingFilterMeta {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastMeta
}

func (p *RodAvitoListingProcessor) FindSimilar(ctx context.Context, listingURL string) ([]SimilarListing, error) {
	if !tokens.ValidListingURL(listingURL) {
		return nil, domain.ErrInvalidListingURL
	}
	if !tokens.IsListingFilterURL(listingURL) {
		return nil, domain.ErrListingSearchRodFilterOnly
	}
	binding, err := p.bindings.PickReady(ctx)
	if err != nil {
		return nil, err
	}

	meta := tokens.ParseListingFilterMeta(listingURL)
	var clientErr error
	if p.client != nil {
		items, err := p.client.FindFromFilterURL(ctx, binding.ProxyEndpoint, listingURL, binding.UserDataDir)
		if err == nil {
			p.mu.Lock()
			p.lastMeta = meta
			p.mu.Unlock()
			return items, nil
		}
		clientErr = err
		if !isAvitoFetchRecoverable(err) {
			return nil, err
		}
	}

	items, htmlMeta, err := p.collectSERP(ctx, binding, listingURL, meta, p.fetch)
	if err == nil {
		p.mu.Lock()
		p.lastMeta = htmlMeta
		p.mu.Unlock()
		return items, nil
	}
	if clientErr != nil {
		return nil, clientErr
	}
	return nil, err
}

func (p *RodAvitoListingProcessor) collectSERP(
	ctx context.Context,
	binding repository.PickReadyBindingRow,
	listingURL string,
	meta tokens.ListingFilterMeta,
	fetch AvitoPageFetcher,
) ([]SimilarListing, tokens.ListingFilterMeta, error) {
	if fetch == nil {
		return nil, meta, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	baseURL := tokens.NormalizeListingSearchURL(listingURL)
	seen := map[string]struct{}{}
	out := make([]SimilarListing, 0, tokens.ListingSearchSERPMaxItems)
	var lastHTML string

	for page := 1; page <= tokens.ListingSearchSERPMaxPages; page++ {
		if page > 1 {
			select {
			case <-ctx.Done():
				return nil, meta, ctx.Err()
			case <-time.After(tokens.RodInterRequestDelay):
			}
		}
		pageURL := tokens.ListingSearchSERPPageURL(baseURL, page)
		html, err := fetch.FetchHTML(ctx, binding.ProxyEndpoint, binding.UserDataDir, pageURL)
		if err != nil {
			return nil, meta, err
		}
		if tokens.IsAvitoChallengeHTML(html) {
			_, _ = p.bindings.MarkSessionChallenge(ctx, binding.ID, tokens.ErrMsgSessionChallenge)
			return nil, meta, newAvitoFetchFailure(tokens.AvitoFetchStateChallenge, 0, true)
		}
		items, err := tokens.ParseAvitoSERPItems(html)
		if err != nil {
			return nil, meta, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
		}
		if len(items) == 0 {
			break
		}
		lastHTML = html
		for _, it := range items {
			if _, ok := seen[it.AvitoID]; ok {
				continue
			}
			seen[it.AvitoID] = struct{}{}
			out = append(out, SimilarListing{AvitoID: it.AvitoID, Title: it.Title})
			if len(out) >= tokens.ListingSearchSERPMaxItems {
				break
			}
		}
		if len(out) >= tokens.ListingSearchSERPMaxItems {
			break
		}
	}

	if lastHTML != "" {
		meta = tokens.MergeFilterMetaFromSERPHTML(meta, lastHTML)
	}
	if len(out) == 0 {
		return nil, meta, newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, 0, false)
	}
	return out, meta, nil
}
