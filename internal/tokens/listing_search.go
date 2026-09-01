package tokens

import (
	"net/url"
	"strconv"
	"strings"
)

// Avito listing URL tokens.
const (
	AvitoListingHostApex    = "avito.ru"
	AvitoListingHostPrimary = "www.avito.ru"
	AvitoURLSchemeHTTP      = "http"
	AvitoURLSchemeHTTPS     = "https"

	FixtureListingSlug1 = "moskva/telefony/iphone_15_1234567890"
	FixtureListingSlug2 = "moskva/telefony/samsung_s24_9876543210"
	FixtureListingID1   = "1234567890"
	FixtureListingID2   = "9876543210"

	FixtureInvalidListingHost = "example.com"
	FixtureInvalidListingPath = "not-avito"

	ListingStubAvitoIDPrefix = "similar"
	ListingIDParseFallback   = "0"

	ListingSearchStubResultCount = 2
)

// AvitoListingHosts are allowed Host values for listing_search query URLs.
var AvitoListingHosts = []string{
	AvitoListingHostApex,
	AvitoListingHostPrimary,
}

var avitoURLSchemes = []string{AvitoURLSchemeHTTP, AvitoURLSchemeHTTPS}

// FixtureListingURL is a contracted valid listing URL for tests.
var FixtureListingURL = AvitoListingURL(AvitoListingHostPrimary, FixtureListingSlug1)

// FixtureListingURL2 is a second valid listing URL (multi-task tests).
var FixtureListingURL2 = AvitoListingURL(AvitoListingHostPrimary, FixtureListingSlug2)

// FixtureInvalidListingURL must be rejected by ValidListingURL.
var FixtureInvalidListingURL = AvitoListingURL(FixtureInvalidListingHost, FixtureInvalidListingPath)

// SchemaListingSearchTables are DB tables introduced for listing_search.
var SchemaListingSearchTables = []string{
	"task_items",
}

// SchemaListingSearchTaskStatuses are task_status enum values for listing_search lifecycle.
var SchemaListingSearchTaskStatuses = []string{
	"COMPLETED",
}

// AvitoListingURL builds a listing URL from host + slug (no leading slash on slug).
func AvitoListingURL(host, slug string) string {
	return AvitoURLSchemeHTTPS + "://" + host + "/" + strings.TrimPrefix(slug, "/")
}

// ValidListingURL reports whether query is an Avito listing URL (listing_search input).
func ValidListingURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	schemeOK := false
	for _, s := range avitoURLSchemes {
		if u.Scheme == s {
			schemeOK = true
			break
		}
	}
	if !schemeOK {
		return false
	}
	host := strings.ToLower(u.Host)
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	for _, h := range AvitoListingHosts {
		if host == h {
			return strings.Trim(u.Path, "/") != ""
		}
	}
	return false
}

// ListingIDFromURL extracts a stable listing id suffix from an Avito URL (stub + future parser).
func ListingIDFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return ListingIDParseFallback
	}
	seg := u.Path
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	parts := strings.Split(seg, "_")
	if len(parts) == 0 {
		if seg == "" {
			return ListingIDParseFallback
		}
		return seg
	}
	last := parts[len(parts)-1]
	if last == "" {
		return ListingIDParseFallback
	}
	return last
}

// StubSimilarAvitoID builds a contracted stub result avito_id.
func StubSimilarAvitoID(listingID string, rank int) string {
	return ListingStubAvitoIDPrefix + "-" + listingID + "-" + strconv.Itoa(rank)
}

// StubSimilarTitle builds a contracted stub result title.
func StubSimilarTitle(listingID string, rank int) string {
	return "Similar listing " + strconv.Itoa(rank) + " for " + listingID
}
