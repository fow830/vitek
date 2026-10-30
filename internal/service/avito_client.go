package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// AvitoClient fetches listing_search data from Avito web endpoints with heal/rotate.
type AvitoClient struct {
	base         string
	client       *http.Client
	changeIPURL  string
	changeIPFunc func(context.Context, string) error
	healBackoff  time.Duration
	maxRotates   int
}

type AvitoClientOption func(*AvitoClient)

func WithAvitoHTTPBase(base string) AvitoClientOption {
	return func(c *AvitoClient) {
		c.base = strings.TrimSuffix(strings.TrimSpace(base), "/")
	}
}

// WithProxyChangeIPURL sets the mobile-proxy change-IP endpoint used on healable blocks.
func WithProxyChangeIPURL(raw string) AvitoClientOption {
	return func(c *AvitoClient) {
		c.changeIPURL = strings.TrimSpace(raw)
	}
}

func NewAvitoClient(opts ...AvitoClientOption) *AvitoClient {
	c := &AvitoClient{
		base: tokens.AvitoHTTPSBase,
		client: &http.Client{
			Timeout: tokens.AvitoHTTPClientTimeout,
		},
		changeIPFunc: ChangeProxyIP,
		healBackoff:  tokens.AvitoRateLimitBackoff,
		maxRotates:   tokens.AvitoHealMaxRotates,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetChangeIPFunc overrides change-IP for tests.
func (c *AvitoClient) SetChangeIPFunc(fn func(context.Context, string) error) {
	c.changeIPFunc = fn
}

// SetHealBackoff overrides heal sleep (0 allowed in tests).
func (c *AvitoClient) SetHealBackoff(d time.Duration) {
	c.healBackoff = d
}

// WithAvitoHealMaxRotates sets extra change-IP rotates after the first GET.
// 0 = single attempt (no rotate); negative values are ignored (keep default).
func WithAvitoHealMaxRotates(n int) AvitoClientOption {
	return func(c *AvitoClient) {
		if n >= 0 {
			c.maxRotates = n
		}
	}
}

type avitoItemDetails struct {
	ID         string
	Title      string
	CategoryID string
	LocationID string
}

func (c *AvitoClient) FindSimilar(ctx context.Context, proxyEndpoint, sourceItemID string) ([]SimilarListing, error) {
	details, err := c.fetchItemDetails(ctx, proxyEndpoint, sourceItemID, "")
	if err != nil {
		return nil, err
	}
	raw, err := c.fetchSearch(ctx, proxyEndpoint, details.CategoryID, details.LocationID, tokens.AvitoSimilarSearchLimit, "")
	if err != nil {
		return nil, err
	}
	out := make([]SimilarListing, 0, len(raw))
	for _, hit := range raw {
		if hit.AvitoID == sourceItemID {
			continue
		}
		out = append(out, hit)
	}
	return out, nil
}

// FindFromFilterURL loads hits from the filter SERP page itself (exact URL params: f=, etc.).
func (c *AvitoClient) FindFromFilterURL(ctx context.Context, proxyEndpoint, filterURL, userDataDir string) ([]SimilarListing, error) {
	if !tokens.IsListingFilterURL(filterURL) {
		return nil, domain.ErrInvalidListingURL
	}
	pageURL := tokens.ListingSearchRewriteHTTPBase(tokens.ListingSearchFilterFetchURL(filterURL), c.base)
	body, err := c.doGETWithHeal(ctx, proxyEndpoint, pageURL, tokens.MIMETextHTML, userDataDir)
	if err != nil {
		return nil, err
	}
	parsed, err := tokens.ParseAvitoSERPItems(string(body))
	if err != nil {
		return nil, newAvitoFetchFailure(tokens.AvitoFetchStateOther, http.StatusOK, false)
	}
	if len(parsed) == 0 {
		return nil, newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, http.StatusOK, false)
	}
	out := make([]SimilarListing, 0, len(parsed))
	for _, it := range parsed {
		out = append(out, SimilarListing{AvitoID: it.AvitoID, Title: it.Title})
	}
	return out, nil
}

func (c *AvitoClient) fetchItemDetails(ctx context.Context, proxyEndpoint, itemID, userDataDir string) (avitoItemDetails, error) {
	body, err := c.doGETWithHeal(ctx, proxyEndpoint, tokens.AvitoItemDetailsURL(c.base, itemID), tokens.AvitoHTTPAccept, userDataDir)
	if err != nil {
		return avitoItemDetails{}, err
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return avitoItemDetails{}, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	id := avitoStringField(raw, tokens.AvitoJSONFieldID)
	if id == "" {
		id = itemID
	}
	title := avitoStringField(raw, tokens.AvitoJSONFieldTitle)
	categoryID := avitoStringField(raw, tokens.AvitoJSONFieldCategoryID)
	locationID := avitoStringField(raw, tokens.AvitoJSONFieldLocationID)
	if categoryID == "" || locationID == "" {
		return avitoItemDetails{}, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	return avitoItemDetails{
		ID:         id,
		Title:      title,
		CategoryID: categoryID,
		LocationID: locationID,
	}, nil
}

func (c *AvitoClient) fetchSearch(ctx context.Context, proxyEndpoint, categoryID, locationID string, limit int, userDataDir string) ([]SimilarListing, error) {
	body, err := c.doGETWithHeal(ctx, proxyEndpoint, tokens.AvitoItemSearchURL(c.base, categoryID, locationID, limit), tokens.AvitoHTTPAccept, userDataDir)
	if err != nil {
		return nil, err
	}
	return c.parseItemsJSON(body)
}

func (c *AvitoClient) parseItemsJSON(body []byte) ([]SimilarListing, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	items, ok := raw[tokens.AvitoJSONFieldItems].([]any)
	if !ok {
		return nil, newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	out := make([]SimilarListing, 0, len(items))
	for _, row := range items {
		obj, ok := row.(map[string]any)
		if !ok {
			continue
		}
		id := avitoStringField(obj, tokens.AvitoJSONFieldID)
		title := avitoStringField(obj, tokens.AvitoJSONFieldTitle)
		if id == "" || title == "" {
			continue
		}
		out = append(out, SimilarListing{AvitoID: id, Title: title})
	}
	if len(out) == 0 {
		return nil, newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, 0, false)
	}
	return out, nil
}

type avitoGETResult struct {
	body   []byte
	status int
	state  string
	err    error
}

func (c *AvitoClient) doGETWithHeal(ctx context.Context, proxyEndpoint, targetURL, accept, userDataDir string) ([]byte, error) {
	max := c.maxRotates
	if max < 0 {
		max = tokens.AvitoHealMaxRotates
	}
	var last avitoGETResult
	for attempt := 0; attempt <= max; attempt++ {
		last = c.doGETOnce(ctx, proxyEndpoint, targetURL, accept, userDataDir)
		if last.err == nil && tokens.AvitoFetchStateIsContent(last.state) {
			return last.body, nil
		}
		if last.err != nil && !isAvitoFetchRecoverable(last.err) {
			return nil, last.err
		}
		healable := tokens.AvitoFetchStateHealable(last.state) || errors.Is(last.err, domain.ErrListingSearchRateLimited)
		if !healable || attempt == max {
			break
		}
		if c.changeIPURL != "" && c.changeIPFunc != nil {
			_ = c.changeIPFunc(ctx, c.changeIPURL)
		}
		backoff := c.healBackoff
		if backoff > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return nil, wrapAvitoGetFailure(last)
}

func (c *AvitoClient) doGETOnce(ctx context.Context, proxyEndpoint, targetURL, accept, userDataDir string) avitoGETResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return avitoGETResult{err: err}
	}
	ua := tokens.AvitoHTTPUserAgent
	if cookie := CookieHeaderFromStorageState(userDataDir, targetURL); cookie != "" {
		ua = tokens.AvitoHTTPBrowserUserAgent
		req.Header.Set(tokens.HeaderCookie, cookie)
	} else if accept == tokens.AvitoHTTPAccept {
		ua = tokens.AvitoHTTPBrowserUserAgent
	}
	req.Header.Set(tokens.HeaderUserAgent, ua)
	req.Header.Set(tokens.HeaderAccept, accept)
	req.Header.Set(tokens.HeaderAcceptLanguage, tokens.AvitoHTTPAcceptLanguage)

	client := c.client
	if strings.TrimSpace(proxyEndpoint) != "" {
		pc, err := HTTPClientViaProxy(proxyEndpoint, c.client.Timeout)
		if err != nil {
			return avitoGETResult{
				state: tokens.AvitoFetchStateOther,
				err:   newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false),
			}
		}
		client = pc
	}

	res, err := client.Do(req)
	if err != nil {
		return avitoGETResult{
			state: tokens.AvitoFetchStateOther,
			err:   newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false),
		}
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(tokens.AvitoHTTPMaxBodyBytes)))
	if err != nil {
		return avitoGETResult{err: err, status: res.StatusCode}
	}
	// JSON API responses are "content" on HTTP 200 even without SERP/MFE markers.
	if res.StatusCode == http.StatusOK && accept == tokens.AvitoHTTPAccept {
		return avitoGETResult{body: body, status: res.StatusCode, state: tokens.AvitoFetchStateContent}
	}
	state := tokens.ClassifyAvitoHTTPResponse(res.StatusCode, string(body))
	if tokens.AvitoFetchStateIsContent(state) {
		return avitoGETResult{body: body, status: res.StatusCode, state: state}
	}
	return avitoGETResult{
		body:   body,
		status: res.StatusCode,
		state:  state,
		err:    errorFromAvitoFetchState(state, res.StatusCode),
	}
}

// ChangeProxyIP hits a mobile-proxy change-IP URL (no-op when empty).
func ChangeProxyIP(ctx context.Context, changeURL string) error {
	changeURL = strings.TrimSpace(changeURL)
	if changeURL == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, changeURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPBrowserUserAgent)
	client := &http.Client{Timeout: tokens.AvitoChangeIPTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, int64(tokens.AvitoHTTPDiscardBodyBytes)))
	if !tokens.HTTPStatusOK(res.StatusCode) {
		return fmt.Errorf("%s: status %d", tokens.ErrMsgProxyChangeIPFailed, res.StatusCode)
	}
	return nil
}

func avitoStringField(obj map[string]any, key string) string {
	return tokens.AvitoMapString(obj, key)
}
