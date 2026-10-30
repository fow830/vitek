package tokens

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Avito HTTP client tokens (listing_search worker).
const (
	AvitoHTTPSBase             = "https://www.avito.ru"
	AvitoWebPathItemDetails    = "/web/1/main/items/"
	AvitoWebPathItemSearch     = "/web/1/main/items"
	AvitoHTTPUserAgent         = "Mozilla/5.0 (compatible; VitekListingSearch/1.0)"
	AvitoHTTPBrowserUserAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	AvitoStorageStateFile      = "storage_state.json"
	AvitoHTTPAccept            = "application/json, text/plain, */*"
	AvitoHTTPAcceptLanguage    = "ru-RU,ru;q=0.9"
	AvitoHTTPClientTimeout     = 30 * time.Second
	AvitoChangeIPTimeout       = 20 * time.Second
	AvitoHTTPMaxBodyBytes      = 4 << 20
	// AvitoHTTPDiscardBodyBytes caps drain of unused response bodies (change-IP, probes).
	AvitoHTTPDiscardBodyBytes = 1 << 20
	AvitoHTTPProbeDiscardBytes = 64 << 10
	AvitoQueryCategoryID       = "categoryId"
	AvitoQueryLocationID       = "locationId"
	AvitoQueryLimit            = "limit"
	AvitoQueryFilterF          = "f"
	AvitoQueryPresentationType = "presentationType"
	AvitoQueryGeoCoords        = "geoCoords"
	AvitoQuerySort             = "sort"
	AvitoQueryLastStamp        = "lastStamp"
	AvitoSortDate              = "date"
	AvitoPresentationTypeSerp  = "serp"
	AvitoSimilarSearchLimit    = 50
	AvitoRateLimitBackoff      = 8 * time.Second
	// AvitoRateLimitMaxRetries aliases AvitoHealMaxRotates (legacy contract name).
	AvitoRateLimitMaxRetries = AvitoHealMaxRotates

	AvitoJSONFieldItems      = "items"
	AvitoJSONFieldID         = "id"
	AvitoJSONFieldTitle      = "title"
	AvitoJSONFieldCategoryID = "categoryId"
	AvitoJSONFieldLocationID = "locationId"

	AvitoSERPAttrItemID      = "data-item-id"
	AvitoSERPMarkerItemTitle = "item-title"
	AvitoSERPMarkerItem      = "item"

	ListingSearchProcessorStub  = "stub"
	ListingSearchProcessorAvito = "avito"
	ListingSearchProcessorRod   = "rod"

	RodPageFetchTimeout = 45 * time.Second
	RodAboutBlankURL    = "about:blank"
	// RodHeadlessDefault false = headed Google Chrome under Xvfb (desktop-like).
	RodHeadlessDefault   = false
	RodChromeXVFBDefault = true
	// RodChromeHeadlessMode is Chrome's --headless= value when headless is forced.
	RodChromeHeadlessMode     = "new"
	RodChromeArgDisableDevShm = "disable-dev-shm-usage"
	RodChromeArgNoSandbox     = "no-sandbox"
	RodChromeArgLang          = "lang"
	RodChromeArgDisableBlink  = "disable-blink-features"
	RodChromeBlinkAutomation  = "AutomationControlled"
	// RodChromeArgForceWebRTCIPHandling blocks host UDP ICE (no server IP leak via STUN).
	RodChromeArgForceWebRTCIPHandling = "force-webrtc-ip-handling-policy"
	RodChromeWebRTCIPHandlingPolicy   = "disable_non_proxied_udp"
	RodChromeArgUserAgent             = "user-agent"
	RodChromeArgWindowSize            = "window-size"
	RodChromeWindowSize               = "1920,1080"
	RodChromeXVFBArgAuto              = "-a"
	RodChromeXVFBArgScreen            = "-s"
	RodChromeXVFBScreen               = "-screen 0 1920x1080x24"
	RodChromeProfileDirName           = "Default"
	RodChromePreferencesFile          = "Preferences"
	RodChromePrefWebRTC               = "webrtc"
	RodChromePrefWebRTCIPHandling     = "ip_handling_policy"
	RodChromePrefWebRTCMultipleRoutes = "multiple_routes_enabled"
	RodChromePrefWebRTCNonProxiedUDP  = "nonproxied_udp_enabled"
	RodChromePrefIntl                 = "intl"
	RodChromePrefSelectedLanguages    = "selected_languages"
	RodChromePrefAcceptLanguages      = "accept_languages"
	RodChromePrefProfile              = "profile"
	RodChromePrefExitType             = "exit_type"
	RodChromePrefExitTypeNormal       = "Normal"
	RodChromePrefExitedCleanly        = "exited_cleanly"
	RodChromeLangFallbackList         = ",ru,en-US,en"
	RodChromeContentPoll              = 250 * time.Millisecond
	RodChromeSessionKeySep            = "|"
	RodChromeUserAgent                = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"
	RodChromePlatform                 = "Win32"
	RodChromeUADataPlatform           = "Windows"
	RodChromeUADataPlatformVersion    = "15.0.0"
	RodChromeUADataArchitecture       = "x86"
	RodChromeUADataBitness            = "64"
	RodChromeUAFullVersion            = "152.0.0.0"
	RodChromeBrandChrome              = "Google Chrome"
	RodChromeBrandChromium            = "Chromium"
	RodChromeBrandNotA                = "Not_A Brand"
	RodChromeBrandChromeMajor         = "152"
	RodChromeBrandNotAVersion         = "24"
	RodChromeTimezone                 = "Europe/Moscow"
	RodChromeAcceptLanguage           = "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7"
	RodChromeLang                     = "ru-RU"
	RodChromeEnvTZ                    = "TZ"
	RodChromeEnvLANG                  = "LANG"
	RodChromeLANGValue                = "ru_RU.UTF-8"
	RodChromeHardwareConcurrency      = 8
	RodChromeDeviceMemoryGB           = 8
	// RodChromeCDPMinimal: fetch HTML via DOM.getOuterHTML + Page.loadEventFired (no Runtime.enable / Eval).
	RodChromeCDPMinimal = true

	FixtureAvitoItemID1        = "7654321098"
	FixtureAvitoItemID2        = "7654321099"
	FixtureAvitoItemTitle1     = "iPhone 15 Pro 256GB"
	FixtureAvitoItemTitle2     = "iPhone 14 128GB"
	FixtureAvitoCategoryID     = "84"
	FixtureAvitoLocationID     = "636736"
	FixtureRodProxyEndpoint    = "socks5://127.0.0.1:9"
	FixtureAvitoHTTPBaseMock   = "http://127.0.0.1:9"
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
	AvitoChallengeTitleIP     = "проблема с IP"
	AvitoChallengeTitleAccess = "Доступ ограничен"
	AvitoChallengeHCaptcha    = "hcaptcha"

	ErrMsgListingSearchNoProxy          = "no active proxy for listing search"
	ErrMsgListingSearchNoAccount        = "no active avito account for listing search"
	ErrMsgListingSearchAvitoFetch       = "avito fetch failed"
	ErrMsgListingSearchRateLimited      = "avito rate limited"
	ErrMsgListingSearchRodFilterOnly    = "rod processor supports filter urls only"
	ErrMsgProxyChangeIPFailed           = "proxy change-ip failed"
	ErrMsgRodLaunch                     = "rod launch failed"
	ErrMsgRodConnect                    = "rod connect failed"
	ErrMsgInvalidListingSearchProcessor = "invalid listing search processor"
)

// RodChromeStealthDeleteFlags are go-rod/puppeteer defaults that scream automation.
var RodChromeStealthDeleteFlags = []string{
	"enable-automation",
	"enable-features",
	"metrics-recording-only",
	"use-mock-keychain",
	"disable-popup-blocking",
	"disable-component-extensions-with-background-pages",
	"disable-default-apps",
	"disable-hang-monitor",
	"disable-prompt-on-repost",
	"disable-sync",
	"disable-client-side-phishing-detection",
	"disable-breakpad",
	"disable-ipc-flooding-protection",
	"disable-renderer-backgrounding",
	"disable-backgrounding-occluded-windows",
	"disable-background-timer-throttling",
	"disable-background-networking",
	"disable-features",
	"disable-site-isolation-trials",
	"force-color-profile",
	"no-startup-window",
	"disable-gpu",
}

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

// AvitoItemPageURL builds a public item card URL from a numeric Avito item id.
func AvitoItemPageURL(itemID string) string {
	id := strings.TrimSpace(itemID)
	if id == "" {
		return ""
	}
	return AvitoHTTPSBase + "/" + id
}

// AvitoFilterSearchURL builds GET /web/1/main/items for a filter/stream query.
func AvitoFilterSearchURL(base, locationID, categoryID string, filterQuery url.Values, limit int) string {
	v := url.Values{}
	v.Set(AvitoQueryCategoryID, categoryID)
	v.Set(AvitoQueryLocationID, locationID)
	v.Set(AvitoQueryLimit, strconv.Itoa(limit))
	v.Set(AvitoQuerySort, AvitoSortDate)
	v.Set(AvitoQueryLastStamp, "0")
	v.Set(AvitoQueryPresentationType, AvitoPresentationTypeSerp)
	if filterQuery != nil {
		if f := strings.TrimSpace(filterQuery.Get(AvitoQueryFilterF)); f != "" {
			v.Set(AvitoQueryFilterF, f)
		}
		if val := strings.TrimSpace(filterQuery.Get(AvitoQueryGeoCoords)); val != "" {
			v.Set(AvitoQueryGeoCoords, val)
		}
	}
	return strings.TrimSuffix(base, "/") + AvitoWebPathItemSearch + "?" + v.Encode()
}

// ListingFilterLocationCategory returns location/category from filter URL meta extras when present.
func ListingFilterLocationCategory(filterURL string) (locationID, categoryID string, ok bool) {
	meta := ParseListingFilterMeta(filterURL)
	locationID = strings.TrimSpace(fmt.Sprint(meta.Extras[AvitoJSONFieldLocationID]))
	categoryID = strings.TrimSpace(fmt.Sprint(meta.Extras[AvitoJSONFieldCategoryID]))
	if locationID == "" || locationID == "<nil>" || categoryID == "" || categoryID == "<nil>" {
		return "", "", false
	}
	// JSON numbers decode as float in extras; Sprint may yield "84" or "84.0".
	if i := strings.IndexByte(locationID, '.'); i > 0 {
		locationID = locationID[:i]
	}
	if i := strings.IndexByte(categoryID, '.'); i > 0 {
		categoryID = categoryID[:i]
	}
	return locationID, categoryID, true
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
// Prefers embedded MFE catalog JSON; falls back to data-item-id card markup.
func ParseAvitoSERPItems(html string) ([]AvitoSERPItem, error) {
	if items := ParseAvitoMFECatalogItems(html); len(items) > 0 {
		return items, nil
	}
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
	if strings.Contains(html, AvitoChallengeMarker) {
		return true
	}
	lower := strings.ToLower(html)
	if strings.Contains(lower, AvitoChallengeHCaptcha) {
		return true
	}
	return strings.Contains(html, AvitoChallengeTitleIP) ||
		strings.Contains(html, AvitoChallengeTitleAccess)
}

// AvitoSERPHTMLReady is true when SERP cards, MFE catalog state, or a challenge page is present.
func AvitoSERPHTMLReady(html string) bool {
	if IsAvitoChallengeHTML(html) {
		return true
	}
	return strings.Contains(html, AvitoSERPAttrItemID+`="`) ||
		strings.Contains(html, AvitoMFEStateAttr)
}

// RodChromeSessionKey identifies a reusable Chrome profile+proxy session.
func RodChromeSessionKey(proxyEndpoint, userDataDir string) string {
	return strings.TrimSpace(proxyEndpoint) + RodChromeSessionKeySep + strings.TrimSpace(userDataDir)
}
