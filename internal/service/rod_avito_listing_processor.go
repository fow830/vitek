package service

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
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
	FetchHTML(ctx context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error)
}

// RodAvitoListingProcessor finds filter SERP listings via Chrome (Rod).
// Item URLs are not supported (ErrListingSearchRodFilterOnly).
type RodAvitoListingProcessor struct {
	bindings *Bindings
	fetch    AvitoPageFetcher

	mu       sync.Mutex
	lastMeta tokens.ListingFilterMeta
}

func NewRodAvitoListingProcessor(pool *pgxpool.Pool, fetch AvitoPageFetcher) *RodAvitoListingProcessor {
	if fetch == nil {
		fetch = NewRodPageFetcher("", "")
	}
	return &RodAvitoListingProcessor{
		bindings: NewBindings(pool),
		fetch:    fetch,
	}
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

	baseURL := tokens.NormalizeListingSearchURL(listingURL)
	meta := tokens.ParseListingFilterMeta(listingURL)
	seen := map[string]struct{}{}
	out := make([]SimilarListing, 0, tokens.ListingSearchSERPMaxItems)
	var lastHTML string

	for page := 1; page <= tokens.ListingSearchSERPMaxPages; page++ {
		if page > 1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(tokens.RodInterRequestDelay):
			}
		}
		pageURL := tokens.ListingSearchSERPPageURL(baseURL, page)
		html, err := p.fetch.FetchHTML(ctx, binding.ProxyEndpoint, binding.UserDataDir, pageURL)
		if err != nil {
			return nil, domain.ErrListingSearchAvitoFetch
		}
		if tokens.IsAvitoChallengeHTML(html) {
			_, _ = p.bindings.MarkSessionChallenge(ctx, binding.ID, tokens.ErrMsgSessionChallenge)
			return nil, domain.ErrListingSearchAvitoFetch
		}
		items, err := tokens.ParseAvitoSERPItems(html)
		if err != nil {
			return nil, domain.ErrListingSearchAvitoFetch
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
	p.mu.Lock()
	p.lastMeta = meta
	p.mu.Unlock()

	if len(out) == 0 {
		return nil, domain.ErrListingSearchAvitoFetch
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

func (f *RodPageFetcher) FetchHTML(ctx context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error) {
	pageURL = tokens.ListingSearchRewriteHTTPBase(pageURL, f.httpBase)
	dir := strings.TrimSpace(userDataDir)
	if dir == "" {
		dir = f.userDataDir
	}
	l := launcher.New().Headless(f.headless).NoSandbox(true)
	bin := strings.TrimSpace(os.Getenv(tokens.EnvRodBrowser))
	if bin == "" {
		bin = tokens.DefaultRodChromeBin
	}
	if bin != "" {
		l = l.Bin(bin)
	}
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("%s: %w", tokens.ErrMsgRodLaunch, err)
		}
		l = l.UserDataDir(dir)
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
