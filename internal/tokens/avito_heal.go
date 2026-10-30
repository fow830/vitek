package tokens

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var avitoMFEStatePattern = regexp.MustCompile(`(?is)<script[^>]*` + regexp.QuoteMeta(AvitoMFEStateAttr) + `[^>]*>(.*?)</script>`)

// Avito heal / anti-bot classification (listing_search fetch path).
const (
	AvitoHTTPStatusPoW            = 439
	AvitoHTTPStatusHeaderTooLarge = 431

	AvitoHealMaxRotates      = 18
	AvitoCookieHeaderMaxBytes = 4096

	AvitoFetchStateContent         = "content"
	AvitoFetchStateEmpty           = "empty"
	AvitoFetchStateChallenge       = "challenge"
	AvitoFetchStateChallengePoW    = "challenge_pow"
	AvitoFetchStateRateLimited     = "rate_limited"
	AvitoFetchStateHeaderTooLarge  = "header_too_large"
	AvitoFetchStateForbidden       = "forbidden"
	AvitoFetchStateOther           = "other"

	AvitoMFEStateAttr     = `data-mfe-state="true"`
	AvitoMFELoaderDataKey = "loaderData"
	AvitoMFEDataKey       = "data"
	AvitoMFECatalogKey    = "catalog"
	AvitoMFEItemsKey      = "items"
	AvitoMFETitleKey      = "title"
	AvitoMFEIDKey         = "id"

	AvitoChallengeTitlePoW = "проверка безопасности"
	AvitoHeaderTooLargeMsg = "Request header fields too large"

	ErrMsgListingSearchAvitoEmpty = "avito empty serp"
)

// AvitoCookieAllowlist is the slim session cookie set for Avito/Qrator fetches.
// Analytics cookies must never be bulk-sent (Qrator HTTP 431).
var AvitoCookieAllowlist = []string{
	"srv_id",
	"f",
	"ft",
	"u",
	"v",
	"i",
	"as",
	"da",
	"cssid",
	"qrator_jsid",
	"qrator_jsr",
	"_csrf",
}

// Fixture HTML bodies for heal classification / MFE parse contracts.
const (
	FixtureAvitoIPBlockHTML = `<!doctype html><html><head><title>Доступ ограничен: проблема с IP</title></head><body></body></html>`
	FixtureAvitoPoWHTML     = `<!doctype html><html><head><title>Доступ ограничен: проверка безопасности</title></head><body></body></html>`
	FixtureAvitoHeaderTooLargeHTML = `QRATOR HTTP 431 - Request header fields too large.`
	FixtureAvitoMFECatalogHTML     = `<!doctype html><html><body>` +
		`<script type="mime/invalid" ` + AvitoMFEStateAttr + `>` +
		`{"loaderData":{"data":{"catalog":{"items":[` +
		`{"id":` + FixtureAvitoItemID1 + `,"title":"` + FixtureAvitoItemTitle1 + `"},` +
		`{"id":` + FixtureAvitoItemID2 + `,"title":"` + FixtureAvitoItemTitle2 + `"}` +
		`]}}}}` +
		`</script></body></html>`
)

// AvitoCookieAllowlistSet is a fast membership map for slim cookie filtering.
func AvitoCookieAllowlistSet() map[string]struct{} {
	out := make(map[string]struct{}, len(AvitoCookieAllowlist))
	for _, name := range AvitoCookieAllowlist {
		out[name] = struct{}{}
	}
	return out
}

// ClassifyAvitoHTTPResponse maps status+body into a fetch state token.
func ClassifyAvitoHTTPResponse(status int, body string) string {
	lower := strings.ToLower(body)
	switch status {
	case AvitoHTTPStatusHeaderTooLarge:
		return AvitoFetchStateHeaderTooLarge
	case AvitoHTTPStatusPoW:
		return AvitoFetchStateChallengePoW
	case http.StatusTooManyRequests:
		return AvitoFetchStateRateLimited
	case http.StatusForbidden:
		return AvitoFetchStateForbidden
	case http.StatusOK:
		if IsAvitoChallengeHTML(body) {
			return AvitoFetchStateChallenge
		}
		if strings.Contains(body, AvitoChallengeTitlePoW) {
			return AvitoFetchStateChallengePoW
		}
		if strings.Contains(body, AvitoSERPAttrItemID+`="`) || strings.Contains(body, AvitoMFEStateAttr) {
			return AvitoFetchStateContent
		}
		return AvitoFetchStateEmpty
	default:
		if strings.Contains(lower, strings.ToLower(AvitoHeaderTooLargeMsg)) {
			return AvitoFetchStateHeaderTooLarge
		}
		if IsAvitoChallengeHTML(body) {
			return AvitoFetchStateChallenge
		}
		return AvitoFetchStateOther
	}
}

// AvitoFetchStateHealable reports states that warrant change-IP / session refresh.
func AvitoFetchStateHealable(state string) bool {
	switch state {
	case AvitoFetchStateRateLimited, AvitoFetchStateChallengePoW,
		AvitoFetchStateHeaderTooLarge, AvitoFetchStateChallenge,
		AvitoFetchStateForbidden:
		return true
	default:
		return false
	}
}

// AvitoFetchStateSoftFailure is true when a watch poll should not auto-PAUSE.
func AvitoFetchStateSoftFailure(state string) bool {
	return AvitoFetchStateHealable(state)
}

// AvitoFetchStateIsContent is true when the body is usable SERP/API content.
func AvitoFetchStateIsContent(state string) bool {
	return state == AvitoFetchStateContent
}

// FormatAvitoFetchError builds a tokenized last_error with state + optional status.
func FormatAvitoFetchError(state string, status int) string {
	if status > 0 {
		return fmt.Sprintf("%s: state=%s status=%d", ErrMsgListingSearchRateLimited, state, status)
	}
	return fmt.Sprintf("%s: state=%s", ErrMsgListingSearchRateLimited, state)
}

// FormatAvitoHardFetchError builds a hard-fail last_error.
func FormatAvitoHardFetchError(state string, status int) string {
	if status > 0 {
		return fmt.Sprintf("%s: state=%s status=%d", ErrMsgListingSearchAvitoFetch, state, status)
	}
	if state == AvitoFetchStateEmpty {
		return fmt.Sprintf("%s: state=%s", ErrMsgListingSearchAvitoEmpty, state)
	}
	return fmt.Sprintf("%s: state=%s", ErrMsgListingSearchAvitoFetch, state)
}

// ParseAvitoMFECatalogItems extracts catalog items from MFE inline JSON state.
func ParseAvitoMFECatalogItems(html string) []AvitoSERPItem {
	matches := avitoMFEStatePattern.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]AvitoSERPItem, 0)
	seen := map[string]struct{}{}
	for _, m := range matches {
		raw := strings.TrimSpace(m[1])
		if raw == "" {
			continue
		}
		var root any
		if err := json.Unmarshal([]byte(raw), &root); err != nil {
			continue
		}
		for _, item := range walkMFECatalogItems(root) {
			if item.AvitoID == "" || item.Title == "" {
				continue
			}
			if _, ok := seen[item.AvitoID]; ok {
				continue
			}
			seen[item.AvitoID] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func walkMFECatalogItems(node any) []AvitoSERPItem {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil
	}
	if loader, ok := obj[AvitoMFELoaderDataKey].(map[string]any); ok {
		obj = loader
	}
	if data, ok := obj[AvitoMFEDataKey].(map[string]any); ok {
		obj = data
	}
	if catalog, ok := obj[AvitoMFECatalogKey].(map[string]any); ok {
		obj = catalog
	}
	items, ok := obj[AvitoMFEItemsKey].([]any)
	if !ok {
		// Nested search: loaderData may wrap differently.
		for _, v := range obj {
			if child, ok := v.(map[string]any); ok {
				if got := walkMFECatalogItems(child); len(got) > 0 {
					return got
				}
			}
		}
		return nil
	}
	out := make([]AvitoSERPItem, 0, len(items))
	for _, row := range items {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		id := AvitoAnyString(m[AvitoMFEIDKey])
		title := AvitoAnyString(m[AvitoMFETitleKey])
		if id == "" || title == "" {
			continue
		}
		out = append(out, AvitoSERPItem{AvitoID: id, Title: title})
	}
	return out
}

// AvitoAnyString coerces JSON scalar fields (string/number) to trimmed text.
func AvitoAnyString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%.0f", t)
	case json.Number:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

// AvitoMapString reads a map field via AvitoAnyString.
func AvitoMapString(obj map[string]any, key string) string {
	if obj == nil {
		return ""
	}
	return AvitoAnyString(obj[key])
}
