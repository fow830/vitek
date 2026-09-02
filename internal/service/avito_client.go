package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"vitek/internal/domain"
	"vitek/internal/tokens"
)

// AvitoClient fetches listing_search data from Avito web JSON endpoints.
type AvitoClient struct {
	base   string
	client *http.Client
}

type AvitoClientOption func(*AvitoClient)

func WithAvitoHTTPBase(base string) AvitoClientOption {
	return func(c *AvitoClient) {
		c.base = strings.TrimSuffix(strings.TrimSpace(base), "/")
	}
}

func NewAvitoClient(opts ...AvitoClientOption) *AvitoClient {
	c := &AvitoClient{
		base: tokens.AvitoHTTPSBase,
		client: &http.Client{
			Timeout: tokens.AvitoHTTPClientTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type avitoItemDetails struct {
	ID         string
	Title      string
	CategoryID string
	LocationID string
}

func (c *AvitoClient) FindSimilar(ctx context.Context, proxyEndpoint, sourceItemID string) ([]SimilarListing, error) {
	details, err := c.fetchItemDetails(ctx, proxyEndpoint, sourceItemID)
	if err != nil {
		return nil, err
	}
	raw, err := c.fetchSearch(ctx, proxyEndpoint, details.CategoryID, details.LocationID, tokens.AvitoSimilarSearchLimit)
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

func (c *AvitoClient) FindFromFilterURL(ctx context.Context, proxyEndpoint, filterURL string) ([]SimilarListing, error) {
	filterURL = tokens.NormalizeListingSearchURL(filterURL)
	u, err := url.Parse(filterURL)
	if err != nil {
		return nil, domain.ErrInvalidListingURL
	}
	pageURL := strings.TrimSuffix(c.base, "/") + u.Path
	if q := u.RawQuery; q != "" {
		pageURL += "?" + q
	}
	htmlBody, err := c.doGET(ctx, proxyEndpoint, pageURL, tokens.MIMETextHTML)
	if err != nil {
		return nil, err
	}
	locationID, categoryID, ok := tokens.ParseAvitoFilterPageIDs(string(htmlBody))
	if !ok {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	apiURL := tokens.AvitoFilterSearchURL(c.base, locationID, categoryID, tokens.ListingSearchFilterQueryForFetch(filterURL), tokens.AvitoSimilarSearchLimit)
	body, err := c.doGET(ctx, proxyEndpoint, apiURL, tokens.AvitoHTTPAccept)
	if err != nil {
		return nil, err
	}
	return c.parseItemsJSON(body)
}

func (c *AvitoClient) fetchItemDetails(ctx context.Context, proxyEndpoint, itemID string) (avitoItemDetails, error) {
	body, err := c.doGET(ctx, proxyEndpoint, tokens.AvitoItemDetailsURL(c.base, itemID), tokens.AvitoHTTPAccept)
	if err != nil {
		return avitoItemDetails{}, err
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return avitoItemDetails{}, domain.ErrListingSearchAvitoFetch
	}
	id := avitoStringField(raw, tokens.AvitoJSONFieldID)
	if id == "" {
		id = itemID
	}
	title := avitoStringField(raw, tokens.AvitoJSONFieldTitle)
	categoryID := avitoStringField(raw, tokens.AvitoJSONFieldCategoryID)
	locationID := avitoStringField(raw, tokens.AvitoJSONFieldLocationID)
	if categoryID == "" || locationID == "" {
		return avitoItemDetails{}, domain.ErrListingSearchAvitoFetch
	}
	return avitoItemDetails{
		ID:         id,
		Title:      title,
		CategoryID: categoryID,
		LocationID: locationID,
	}, nil
}

func (c *AvitoClient) fetchSearch(ctx context.Context, proxyEndpoint, categoryID, locationID string, limit int) ([]SimilarListing, error) {
	body, err := c.doGET(ctx, proxyEndpoint, tokens.AvitoItemSearchURL(c.base, categoryID, locationID, limit), tokens.AvitoHTTPAccept)
	if err != nil {
		return nil, err
	}
	return c.parseItemsJSON(body)
}

func (c *AvitoClient) parseItemsJSON(body []byte) ([]SimilarListing, error) {
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	items, ok := raw[tokens.AvitoJSONFieldItems].([]any)
	if !ok {
		return nil, domain.ErrListingSearchAvitoFetch
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
		return nil, domain.ErrListingSearchAvitoFetch
	}
	return out, nil
}

func (c *AvitoClient) doGET(ctx context.Context, proxyEndpoint, targetURL, accept string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(tokens.HeaderUserAgent, tokens.AvitoHTTPUserAgent)
	req.Header.Set(tokens.HeaderAccept, accept)
	req.Header.Set(tokens.HeaderAcceptLanguage, tokens.AvitoHTTPAcceptLanguage)

	client := c.client
	if strings.TrimSpace(proxyEndpoint) != "" {
		pc, err := HTTPClientViaProxy(proxyEndpoint, c.client.Timeout)
		if err != nil {
			return nil, domain.ErrListingSearchAvitoFetch
		}
		client = pc
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, int64(tokens.AvitoHTTPMaxBodyBytes)))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, domain.ErrListingSearchAvitoFetch
	}
	return body, nil
}

func avitoStringField(obj map[string]any, key string) string {
	v, ok := obj[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%.0f", t)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", t)
	}
}
