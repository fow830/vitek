package tokens

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Avito HTTP client tokens (listing_search worker).
const (
	AvitoHTTPSBase          = "https://www.avito.ru"
	AvitoWebPathItemDetails = "/web/1/main/items/"
	AvitoWebPathItemSearch  = "/web/1/main/items"
	AvitoHTTPUserAgent      = "Mozilla/5.0 (compatible; VitekListingSearch/1.0)"
	AvitoHTTPAccept         = "application/json, text/plain, */*"
	AvitoHTTPAcceptLanguage = "ru-RU,ru;q=0.9"
	AvitoHTTPClientTimeout  = 30 * time.Second
	AvitoHTTPMaxBodyBytes   = 4 << 20
	AvitoQueryCategoryID    = "categoryId"
	AvitoQueryLocationID    = "locationId"
	AvitoQueryLimit         = "limit"
	AvitoQueryFilterF       = "f"
	AvitoQueryPresentationType = "presentationType"
	AvitoQueryGeoCoords        = "geoCoords"
	AvitoPresentationTypeSerp  = "serp"
	AvitoSimilarSearchLimit = 50

	AvitoJSONFieldItems      = "items"
	AvitoJSONFieldID         = "id"
	AvitoJSONFieldTitle      = "title"
	AvitoJSONFieldCategoryID = "categoryId"
	AvitoJSONFieldLocationID = "locationId"

	AvitoSERPAttrItemID       = "data-item-id"
	AvitoSERPMarkerItemTitle  = "item-title"
	AvitoSERPMarkerItem       = "item"

	ListingSearchProcessorStub  = "stub"
	ListingSearchProcessorAvito = "avito"
	ListingSearchProcessorRod   = "rod"

	RodPageFetchTimeout = 45 * time.Second
	RodHeadlessDefault  = true

	FixtureAvitoItemID1    = "7654321098"
	FixtureAvitoItemID2    = "7654321099"
	FixtureAvitoItemTitle1 = "iPhone 15 Pro 256GB"
	FixtureAvitoItemTitle2 = "Filter hit 2"
	FixtureAvitoCategoryID = "84"
	FixtureAvitoLocationID = "636736"
	FixtureRodProxyEndpoint  = "socks5://127.0.0.1:9"
	FixtureAvitoHTTPBaseMock = "http://127.0.0.1:9"
	FixtureAvitoFilterPageHTML = `<!doctype html><html><body>` +
		`"locationId":` + FixtureAvitoLocationID + `,"categoryId":` + FixtureAvitoCategoryID +
		`</body></html>`
	FixtureAvitoFilterSERPHTML = `<!doctype html><html><body>` +
		`"locationId":` + FixtureAvitoLocationID + `,"categoryId":` + FixtureAvitoCategoryID +
		`<div ` + AvitoSERPAttrItemID + `="` + FixtureAvitoItemID1 + `" data-marker="` + AvitoSERPMarkerItem + `"><h3 data-marker="` + AvitoSERPMarkerItemTitle + `">` + FixtureAvitoItemTitle1 + `</h3></div>` +
		`<div ` + AvitoSERPAttrItemID + `="` + FixtureAvitoItemID2 + `" data-marker="` + AvitoSERPMarkerItem + `"><h3 data-marker="` + AvitoSERPMarkerItemTitle + `">` + FixtureAvitoItemTitle2 + `</h3></div>` +
		`</body></html>`
	FixtureAvitoChallengeHTML = `<!doctype html><html><body><div id="challenge-stage">firewall</div></body></html>`
	AvitoChallengeMarker      = "challenge-stage"

	ErrMsgListingSearchNoProxy            = "no active proxy for listing search"
	ErrMsgListingSearchNoAccount          = "no active avito account for listing search"
	ErrMsgListingSearchAvitoFetch         = "avito fetch failed"
	ErrMsgListingSearchRodFilterOnly      = "rod processor supports filter urls only"
	ErrMsgRodLaunch                       = "rod launch failed"
	ErrMsgRodConnect                      = "rod connect failed"
	ErrMsgInvalidListingSearchProcessor   = "invalid listing search processor"
)

// ListingSearchProcessors are allowed LISTING_SEARCH_PROCESSOR values.
var ListingSearchProcessors = []string{
	ListingSearchProcessorStub,
	ListingSearchProcessorAvito,
	ListingSearchProcessorRod,
}

// SchemaAvitoSecretsTables are DB tables for Avito credentials.
var SchemaAvitoSecretsTables = []string{
	"avito_account_secrets",
}

// AvitoItemDetailsURL builds GET item details URL for a numeric Avito item id.
func AvitoItemDetailsURL(base, itemID string) string {
	return strings.TrimSuffix(base, "/") + AvitoWebPathItemDetails + itemID
}

// AvitoFilterSearchURL builds GET /web/1/main/items for a filter/stream query.
func AvitoFilterSearchURL(base, locationID, categoryID string, filterQuery url.Values, limit int) string {
	v := url.Values{}
	v.Set(AvitoQueryCategoryID, categoryID)
	v.Set(AvitoQueryLocationID, locationID)
	v.Set(AvitoQueryLimit, strconv.Itoa(limit))
	v.Set(AvitoQueryPresentationType, AvitoPresentationTypeSerp)
	if f := strings.TrimSpace(filterQuery.Get(AvitoQueryFilterF)); f != "" {
		v.Set(AvitoQueryFilterF, f)
	}
	if val := strings.TrimSpace(filterQuery.Get(AvitoQueryGeoCoords)); val != "" {
		v.Set(AvitoQueryGeoCoords, val)
	}
	return strings.TrimSuffix(base, "/") + AvitoWebPathItemSearch + "?" + v.Encode()
}

// AvitoItemSearchURL builds GET similar search URL with query params.
func AvitoItemSearchURL(base, categoryID, locationID string, limit int) string {
	v := url.Values{}
	v.Set(AvitoQueryCategoryID, categoryID)
	v.Set(AvitoQueryLocationID, locationID)
	v.Set(AvitoQueryLimit, strconv.Itoa(limit))
	return strings.TrimSuffix(base, "/") + AvitoWebPathItemSearch + "?" + v.Encode()
}

// AvitoSERPItem is a listing card parsed from filter SERP HTML (Rod path).
type AvitoSERPItem struct {
	AvitoID string
	Title   string
}

var avitoSERPItemBlockPattern = regexp.MustCompile(
	`(?is)` + AvitoSERPAttrItemID + `="(\d+)"[^>]*>.*?data-marker="` + regexp.QuoteMeta(AvitoSERPMarkerItemTitle) + `"[^>]*>([^<]+)<`,
)

// ParseAvitoSERPItems extracts listing cards from Avito filter SERP HTML.
func ParseAvitoSERPItems(html string) ([]AvitoSERPItem, error) {
	matches := avitoSERPItemBlockPattern.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	out := make([]AvitoSERPItem, 0, len(matches))
	seen := map[string]struct{}{}
	for _, m := range matches {
		id := strings.TrimSpace(m[1])
		title := strings.TrimSpace(m[2])
		if id == "" || title == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, AvitoSERPItem{AvitoID: id, Title: title})
	}
	return out, nil
}

// ListingSearchRewriteHTTPBase rewrites an Avito page URL onto httpBase host (Rod/mock).
func ListingSearchRewriteHTTPBase(pageURL, httpBase string) string {
	base := strings.TrimSuffix(strings.TrimSpace(httpBase), "/")
	if base == "" || base == AvitoHTTPSBase {
		return pageURL
	}
	u, err := url.Parse(pageURL)
	if err != nil {
		return pageURL
	}
	b, err := url.Parse(base)
	if err != nil || b.Host == "" {
		return pageURL
	}
	u.Scheme = b.Scheme
	u.Host = b.Host
	return u.String()
}

// ListingSearchSERPPageURL sets/replaces SERP page query param (1-based; page 1 omits param).
func ListingSearchSERPPageURL(raw string, page int) string {
	if page <= 1 {
		return raw
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	q := u.Query()
	q.Set(ListingSearchSERPPageParam, strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// MergeFilterMetaFromSERPHTML enriches URL meta with location/category ids from SERP HTML.
func MergeFilterMetaFromSERPHTML(base ListingFilterMeta, html string) ListingFilterMeta {
	if loc, cat, ok := ParseAvitoFilterPageIDs(html); ok {
		if base.Extras == nil {
			base.Extras = map[string]any{}
		}
		base.Extras[AvitoJSONFieldLocationID] = loc
		base.Extras[AvitoJSONFieldCategoryID] = cat
	}
	return base
}

// ProxyProbeURL is the contracted health probe target (Avito origin).
func ProxyProbeURL() string {
	return AvitoHTTPSBase + ProxyProbePath
}

// IsAvitoChallengeHTML reports Avito anti-bot / challenge interstitial HTML.
func IsAvitoChallengeHTML(html string) bool {
	return strings.Contains(html, AvitoChallengeMarker)
}
