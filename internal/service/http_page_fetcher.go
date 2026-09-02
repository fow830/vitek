package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// HTTPPageFetcher loads Avito HTML via SOCKS/HTTP proxy (SERP channel without Chrome).
type HTTPPageFetcher struct {
	httpBase string
	timeout  time.Duration
}

func NewHTTPPageFetcher(httpBase string) *HTTPPageFetcher {
	return &HTTPPageFetcher{
		httpBase: strings.TrimSpace(httpBase),
		timeout:  tokens.AvitoHTTPClientTimeout,
	}
}

func (f *HTTPPageFetcher) FetchHTML(ctx context.Context, proxyEndpoint, _ /* userDataDir */, pageURL string) (string, error) {
	pageURL = tokens.ListingSearchRewriteHTTPBase(pageURL, f.httpBase)
	client, err := HTTPClientViaProxy(proxyEndpoint, f.timeout)
	if err != nil {
		return "", domain.ErrListingSearchAvitoFetch
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPUserAgent)
	req.Header.Set(tokens.HeaderAccept, tokens.MIMETextHTML)
	req.Header.Set(tokens.HeaderAcceptLanguage, tokens.AvitoHTTPAcceptLanguage)

	res, err := client.Do(req)
	if err != nil {
		return "", domain.ErrListingSearchAvitoFetch
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(tokens.AvitoHTTPMaxBodyBytes)))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", domain.ErrListingSearchAvitoFetch
	}
	html := string(body)
	if strings.TrimSpace(html) == "" {
		return "", domain.ErrListingSearchAvitoFetch
	}
	return html, nil
}
