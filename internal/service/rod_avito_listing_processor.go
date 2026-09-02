package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jackc/pgx/v5/pgxpool"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// AvitoPageFetcher loads an Avito page HTML (browser session).
type AvitoPageFetcher interface {
	FetchHTML(ctx context.Context, proxyEndpoint, pageURL string) (string, error)
}

// RodAvitoListingProcessor finds filter SERP listings via Chrome (Rod).
// Item URLs are not supported (ErrListingSearchRodFilterOnly).
type RodAvitoListingProcessor struct {
	proxies *Proxies
	avito   *AvitoAccounts
	fetch   AvitoPageFetcher
}

func NewRodAvitoListingProcessor(pool *pgxpool.Pool, fetch AvitoPageFetcher) *RodAvitoListingProcessor {
	if fetch == nil {
		fetch = NewRodPageFetcher("", "")
	}
	return &RodAvitoListingProcessor{
		proxies: NewProxies(pool),
		avito:   NewAvitoAccounts(pool),
		fetch:   fetch,
	}
}

func (p *RodAvitoListingProcessor) FindSimilar(ctx context.Context, listingURL string) ([]SimilarListing, error) {
	if !tokens.ValidListingURL(listingURL) {
		return nil, domain.ErrInvalidListingURL
	}
	if !tokens.IsListingFilterURL(listingURL) {
		return nil, domain.ErrListingSearchRodFilterOnly
	}
	proxies, err := p.proxies.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return nil, domain.ErrListingSearchNoProxy
	}
	// Account presence is a contracted gate (credentials used when session login is contracted).
	if _, err := p.avito.PickActive(ctx); err != nil {
		return nil, domain.ErrListingSearchNoAccount
	}
	proxy := proxies[0].Endpoint
	pageURL := tokens.NormalizeListingSearchURL(listingURL)
	html, err := p.fetch.FetchHTML(ctx, proxy, pageURL)
	if err != nil {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	items, err := tokens.ParseAvitoSERPItems(html)
	if err != nil {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	if len(items) == 0 {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	out := make([]SimilarListing, 0, len(items))
	for _, it := range items {
		out = append(out, SimilarListing{AvitoID: it.AvitoID, Title: it.Title})
	}
	return out, nil
}

// RodPageFetcher opens pages with go-rod Chrome (optional user-data-dir + HTTP base rewrite).
type RodPageFetcher struct {
	userDataDir string
	httpBase    string
	timeout     time.Duration
	headless    bool
}

func NewRodPageFetcher(userDataDir, httpBase string) *RodPageFetcher {
	return &RodPageFetcher{
		userDataDir: strings.TrimSpace(userDataDir),
		httpBase:    strings.TrimSpace(httpBase),
		timeout:     tokens.RodPageFetchTimeout,
		headless:    tokens.RodHeadlessDefault,
	}
}

func (f *RodPageFetcher) FetchHTML(ctx context.Context, proxyEndpoint, pageURL string) (string, error) {
	pageURL = tokens.ListingSearchRewriteHTTPBase(pageURL, f.httpBase)
	l := launcher.New().Headless(f.headless)
	if f.userDataDir != "" {
		l = l.UserDataDir(f.userDataDir)
	}
	if strings.TrimSpace(proxyEndpoint) != "" {
		proxyURL, err := url.Parse(proxyEndpoint)
		if err != nil {
			return "", err
		}
		l = l.Proxy(proxyURL.String())
	}
	controlURL, err := l.Launch()
	if err != nil {
		return "", fmt.Errorf("%s: %w", tokens.ErrMsgRodLaunch, err)
	}
	browser := rod.New().ControlURL(controlURL).Context(ctx)
	if err := browser.Connect(); err != nil {
		return "", fmt.Errorf("%s: %w", tokens.ErrMsgRodConnect, err)
	}
	defer func() { _ = browser.Close() }()

	page, err := browser.Page(proto.TargetCreateTarget{URL: pageURL})
	if err != nil {
		return "", err
	}
	page = page.Context(ctx).Timeout(f.timeout)
	if err := page.WaitLoad(); err != nil {
		return "", domain.ErrListingSearchAvitoFetch
	}
	html, err := page.HTML()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(html) == "" {
		return "", domain.ErrListingSearchAvitoFetch
	}
	return html, nil
}
