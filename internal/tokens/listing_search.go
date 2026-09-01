package tokens

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Avito listing URL tokens.
const (
	AvitoListingHostApex    = "avito.ru"
	AvitoListingHostPrimary = "www.avito.ru"
	AvitoListingHostMobile  = "m.avito.ru"
	AvitoURLSchemeHTTP      = "http"
	AvitoURLSchemeHTTPS     = "https"

	FixtureListingSlug1 = "moskva/telefony/iphone_15_1234567890"
	FixtureListingSlug2 = "moskva/telefony/samsung_s24_9876543210"
	FixtureListingID1   = "1234567890"
	FixtureListingID2   = "9876543210"

	FixtureMobileListingSlug = "sankt-peterburg/telefony/iphone_15_1234567890"
	FixtureFilterListingSlug = "sankt-peterburg/telefony/mobilnye_telefony/apple-ASgBAgICAkS0wa3OqzmwwQ2I_Dc"
	FixtureFilterQueryF      = "ASgBAQECAkS0wA3OqzmwwQ2I_DcDQLLADUSSn~0R1qHtEcqxjBXMsYwV5uANNPbBXPjBXPrBXOjrDjT6_dsC_P3bAv792wIBRcaaDBV7ImZyb20iOjAsInRvIjo3MDAwMH0"

	ListingSearchQueryKindItem   = "item"
	ListingSearchQueryKindFilter = "filter"

	ListingIDMinDigits = 5

	FixtureInvalidListingHost = "example.com"
	FixtureInvalidListingPath = "not-avito"

	ListingStubAvitoIDPrefix = "similar"
	ListingIDParseFallback   = "0"

	ListingSearchStubResultCount = 2

	ListingSearchTaskListLimit int32 = 20

	ListingSearchPollMaxAttempts = 30
	ListingSearchPollIntervalMs  = 500
	ListingSearchWatchPollIntervalMs = 60000

	ListingSearchKindWatch = "watch"
	ListingSearchKindTask  = "task"

	ListingWatchStatusActive = "ACTIVE"
)

// TaskStatus* mirror PostgreSQL task_status enum (SoT for JSON + generated JS).
const (
	TaskStatusPending   = "PENDING"
	TaskStatusRunning   = "RUNNING"
	TaskStatusPaused    = "PAUSED"
	TaskStatusFailed    = "FAILED"
	TaskStatusCompleted = "COMPLETED"
)

// TaskStatusTerminal stops listing_search poll UI in app face JS.
var TaskStatusTerminal = []string{TaskStatusCompleted, TaskStatusFailed}

// AvitoListingHosts are allowed Host values for listing_search query URLs.
var AvitoListingHosts = []string{
	AvitoListingHostApex,
	AvitoListingHostPrimary,
	AvitoListingHostMobile,
}

var avitoURLSchemes = []string{AvitoURLSchemeHTTP, AvitoURLSchemeHTTPS}

// FixtureListingURL is a contracted valid listing URL for tests.
var FixtureListingURL = AvitoListingURL(AvitoListingHostPrimary, FixtureListingSlug1)

// FixtureListingURL2 is a second valid listing URL (multi-task tests).
var FixtureListingURL2 = AvitoListingURL(AvitoListingHostPrimary, FixtureListingSlug2)

// FixtureMobileListingURL is a contracted valid mobile Avito listing URL.
var FixtureMobileListingURL = AvitoListingURL(AvitoListingHostMobile, FixtureMobileListingSlug)

// FixtureFilterListingURL is a contracted Avito filter/stream URL (SERP).
var FixtureFilterListingURL = AvitoListingURL(AvitoListingHostPrimary, FixtureFilterListingSlug) +
	"?presentationType=serp&f=" + url.QueryEscape(FixtureFilterQueryF)

// FixtureMobileFilterListingURL is the same filter on m.avito.ru (normalized to www on task create).
var FixtureMobileFilterListingURL = AvitoListingURL(AvitoListingHostMobile, FixtureFilterListingSlug) +
	"?presentationType=serp&f=" + url.QueryEscape(FixtureFilterQueryF)

// FixtureInvalidListingURL must be rejected by ValidListingURL.
var FixtureInvalidListingURL = AvitoListingURL(FixtureInvalidListingHost, FixtureInvalidListingPath)

// SchemaListingSearchTables are DB tables introduced for listing_search.
var SchemaListingSearchTables = []string{
	"task_items",
	"listing_filter_seen",
	"listing_filter_watches",
	"listing_watch_hits",
}

// SchemaListingWatchStatuses mirror PostgreSQL listing_watch_status enum.
var SchemaListingWatchStatuses = []string{
	ListingWatchStatusActive,
}

// SchemaListingSearchTaskStatuses are task_status enum values added for listing_search.
var SchemaListingSearchTaskStatuses = []string{
	TaskStatusCompleted,
}

// AvitoListingURL builds a listing URL from host + slug (no leading slash on slug).
func AvitoListingURL(host, slug string) string {
	return AvitoURLSchemeHTTPS + "://" + host + "/" + strings.TrimPrefix(slug, "/")
}

// ValidListingURL reports whether query is an Avito listing or filter URL (listing_search input).
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

// IsListingItemURL reports whether the URL points at a single Avito item card.
func IsListingItemURL(raw string) bool {
	return ValidListingURL(raw) && ValidListingID(ListingIDFromURL(raw))
}

// IsListingFilterURL reports whether the URL is an Avito filter/stream (SERP) page.
func IsListingFilterURL(raw string) bool {
	return ValidListingURL(raw) && !IsListingItemURL(raw)
}

// ListingSearchQueryKind classifies listing_search task input.
func ListingSearchQueryKind(raw string) string {
	if IsListingItemURL(raw) {
		return ListingSearchQueryKindItem
	}
	if IsListingFilterURL(raw) {
		return ListingSearchQueryKindFilter
	}
	return ""
}

// NormalizeListingSearchURL canonicalizes host for upstream Avito requests.
func NormalizeListingSearchURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	host := strings.ToLower(u.Host)
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	if host == AvitoListingHostMobile {
		u.Host = AvitoListingHostPrimary
	}
	return u.String()
}

// ValidListingID reports whether id looks like an Avito numeric item id.
func ValidListingID(id string) bool {
	if len(id) < ListingIDMinDigits {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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

// ListingStubKeyFromURL returns a stable stub key for item or filter URLs.
func ListingStubKeyFromURL(raw string) string {
	if id := ListingIDFromURL(raw); ValidListingID(id) {
		return id
	}
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ListingIDParseFallback
	}
	if f := u.Query().Get(AvitoQueryFilterF); f != "" {
		if len(f) > 16 {
			return f[:16]
		}
		return f
	}
	seg := u.Path
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	if seg == "" {
		return ListingIDParseFallback
	}
	return seg
}

var (
	avitoHTMLLocationIDPattern = regexp.MustCompile(`locationId["':]+(\d+)`)
	avitoHTMLCategoryIDPattern = regexp.MustCompile(`categoryId["':]+(\d+)`)
)

// ListingSearchDedupVolatileQueryKeys are dropped when canonicalizing filter URLs.
var ListingSearchDedupVolatileQueryKeys = []string{
	"context",
	"geoCoords",
	"radius",
	"moreExpensive",
}

// ListingSearchPathFilterTokenPrefix marks Avito filter blob suffix in SERP path slugs.
const ListingSearchPathFilterTokenPrefix = "ASgB"

// CanonicalListingSearchQuery returns a stable task/dedup key (host + path; filter query stripped).
func CanonicalListingSearchQuery(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(NormalizeListingSearchURL(raw))
	if err != nil {
		return raw
	}
	u.Host = AvitoListingHostPrimary
	u.RawQuery = ""
	u.Fragment = ""
	path := strings.TrimSuffix(u.Path, "/")
	if path == "" {
		path = "/"
	}
	return AvitoURLSchemeHTTPS + "://" + AvitoListingHostPrimary + path
}

// ListingSearchFilterFFromURL returns Avito filter blob from query or path slug.
func ListingSearchFilterFFromURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(NormalizeListingSearchURL(raw))
	if err != nil {
		return ""
	}
	if f := strings.TrimSpace(u.Query().Get(AvitoQueryFilterF)); f != "" {
		return f
	}
	return ListingSearchFilterFFromPath(u.Path)
}

// ListingSearchFilterFFromPath extracts ASgB… token from a filter SERP path slug.
func ListingSearchFilterFFromPath(path string) string {
	seg := path
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	if idx := strings.Index(seg, "-"+ListingSearchPathFilterTokenPrefix); idx >= 0 {
		return seg[idx+1:]
	}
	if strings.HasPrefix(seg, ListingSearchPathFilterTokenPrefix) {
		return seg
	}
	return ""
}

// ListingSearchFilterQueryForFetch builds query params forwarded to Avito items API.
func ListingSearchFilterQueryForFetch(raw string) url.Values {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(NormalizeListingSearchURL(raw))
	if err != nil {
		return url.Values{}
	}
	out := url.Values{}
	if f := ListingSearchFilterFFromURL(raw); f != "" {
		out.Set(AvitoQueryFilterF, f)
	}
	if pt := strings.TrimSpace(u.Query().Get(AvitoQueryPresentationType)); pt != "" {
		out.Set(AvitoQueryPresentationType, pt)
	}
	return out
}

// ParseAvitoFilterPageIDs extracts location/category ids embedded in a filter SERP HTML page.
func ParseAvitoFilterPageIDs(html string) (locationID, categoryID string, ok bool) {
	loc := avitoHTMLLocationIDPattern.FindStringSubmatch(html)
	cat := avitoHTMLCategoryIDPattern.FindStringSubmatch(html)
	if len(loc) != 2 || len(cat) != 2 {
		return "", "", false
	}
	return loc[1], cat[1], true
}

// StubSimilarAvitoID builds a contracted stub result avito_id.
func StubSimilarAvitoID(listingID string, rank int) string {
	return ListingStubAvitoIDPrefix + "-" + listingID + "-" + strconv.Itoa(rank)
}

// StubSimilarTitle builds a contracted stub result title.
func StubSimilarTitle(listingID string, rank int) string {
	return "Similar listing " + strconv.Itoa(rank) + " for " + listingID
}
