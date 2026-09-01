package tokens

import (
	"net/url"
	"strconv"
	"strings"
)

// Avito HTTP client tokens (listing_search worker).
const (
	AvitoHTTPSBase          = "https://www.avito.ru"
	AvitoWebPathItemDetails = "/web/1/main/items/"
	AvitoWebPathItemSearch  = "/web/1/main/items"
	AvitoHTTPUserAgent      = "Mozilla/5.0 (compatible; VitekListingSearch/1.0)"
	AvitoHTTPAccept         = "application/json, text/plain, */*"
	AvitoHTTPAcceptLanguage = "ru-RU,ru;q=0.9"
	AvitoQueryCategoryID    = "categoryId"
	AvitoQueryLocationID    = "locationId"
	AvitoQueryLimit         = "limit"
	AvitoQueryFilterF       = "f"
	AvitoQueryPresentationType = "presentationType"
	AvitoPresentationTypeSerp  = "serp"
	AvitoSimilarSearchLimit = 10

	AvitoJSONFieldItems      = "items"
	AvitoJSONFieldID         = "id"
	AvitoJSONFieldTitle      = "title"
	AvitoJSONFieldCategoryID = "categoryId"
	AvitoJSONFieldLocationID = "locationId"

	AvitoAccountLoginField = JSONFieldExternalRef

	ListingSearchProcessorStub  = "stub"
	ListingSearchProcessorAvito = "avito"

	FixtureAvitoItemID1    = "7654321098"
	FixtureAvitoItemID2    = "7654321099"
	FixtureAvitoItemTitle1 = "iPhone 15 Pro 256GB"
	FixtureAvitoCategoryID = "84"
	FixtureAvitoLocationID = "636736"
	FixtureAvitoFilterPageHTML = `<!doctype html><html><body>` +
		`"locationId":` + FixtureAvitoLocationID + `,"categoryId":` + FixtureAvitoCategoryID +
		`</body></html>`

	ErrMsgListingSearchNoProxy            = "no active proxy for listing search"
	ErrMsgListingSearchNoAccount          = "no active avito account for listing search"
	ErrMsgListingSearchAvitoFetch         = "avito fetch failed"
	ErrMsgInvalidListingSearchProcessor   = "invalid listing search processor"
)

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
	for _, key := range []string{"geoCoords", "moreExpensive", "context"} {
		if val := strings.TrimSpace(filterQuery.Get(key)); val != "" {
			v.Set(key, val)
		}
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
